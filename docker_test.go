package docker

import (
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/dchest/uniuri"
)

func TestCommandBuild(t *testing.T) {
	tempTag := strings.ToLower(uniuri.New())
	tcs := []struct {
		name  string
		build Build
		want  *exec.Cmd
	}{
		{
			name: "secret from env var",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				TempTag:    tempTag,
				Dockerfile: "Dockerfile",
				Context:    ".",
				SecretEnvs: []string{
					"foo_secret=FOO_SECRET_ENV_VAR",
				},
			},
			want: exec.Command(
				dockerExe,
				"build",
				"--rm=true",
				"-f",
				"Dockerfile",
				"-t",
				tempTag,
				".",
				"--secret id=foo_secret,env=FOO_SECRET_ENV_VAR",
			),
		},
		{
			name: "secret from file",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				TempTag:    tempTag,
				Dockerfile: "Dockerfile",
				Context:    ".",
				SecretFiles: []string{
					"foo_secret=/path/to/foo_secret",
				},
			},
			want: exec.Command(
				dockerExe,
				"build",
				"--rm=true",
				"-f",
				"Dockerfile",
				"-t",
				tempTag,
				".",
				"--secret id=foo_secret,src=/path/to/foo_secret",
			),
		},
		{
			name: "multiple mixed secrets",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				TempTag:    tempTag,
				Dockerfile: "Dockerfile",
				Context:    ".",
				SecretEnvs: []string{
					"foo_secret=FOO_SECRET_ENV_VAR",
					"bar_secret=BAR_SECRET_ENV_VAR",
				},
				SecretFiles: []string{
					"foo_secret=/path/to/foo_secret",
					"bar_secret=/path/to/bar_secret",
				},
			},
			want: exec.Command(
				dockerExe,
				"build",
				"--rm=true",
				"-f",
				"Dockerfile",
				"-t",
				tempTag,
				".",
				"--secret id=foo_secret,env=FOO_SECRET_ENV_VAR",
				"--secret id=bar_secret,env=BAR_SECRET_ENV_VAR",
				"--secret id=foo_secret,src=/path/to/foo_secret",
				"--secret id=bar_secret,src=/path/to/bar_secret",
			),
		},
		{
			name: "invalid mixed secrets",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				TempTag:    tempTag,
				Dockerfile: "Dockerfile",
				Context:    ".",
				SecretEnvs: []string{
					"foo_secret=",
					"=FOO_SECRET_ENV_VAR",
					"",
				},
				SecretFiles: []string{
					"foo_secret=",
					"=/path/to/bar_secret",
					"",
				},
			},
			want: exec.Command(
				dockerExe,
				"build",
				"--rm=true",
				"-f",
				"Dockerfile",
				"-t",
				tempTag,
				".",
			),
		},
		{
			name: "platform argument",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				TempTag:    tempTag,
				Dockerfile: "Dockerfile",
				Context:    ".",
				Platform:   "test/platform",
			},
			want: exec.Command(
				dockerExe,
				"build",
				"--rm=true",
				"-f",
				"Dockerfile",
				"-t",
				tempTag,
				".",
				"--platform",
				"test/platform",
			),
		},
		{
			name: "ssh agent",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				TempTag:    tempTag,
				Dockerfile: "Dockerfile",
				Context:    ".",
				SSHKeyPath: "id_rsa=/root/.ssh/id_rsa",
			},
			want: exec.Command(
				dockerExe,
				"build",
				"--rm=true",
				"-f",
				"Dockerfile",
				"-t",
				tempTag,
				".",
				"--ssh id_rsa=/root/.ssh/id_rsa",
			),
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := commandBuild(tc.build)

			if !reflect.DeepEqual(cmd.String(), tc.want.String()) {
				t.Errorf("Got cmd %v, want %v", cmd, tc.want)
			}
		})
	}
}

func TestGetProxyValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "lowercase env var set",
			key:      "http_proxy",
			envVars:  map[string]string{"http_proxy": "http://proxy:8080"},
			expected: "http://proxy:8080",
		},
		{
			name:     "uppercase env var set",
			key:      "http_proxy",
			envVars:  map[string]string{"HTTP_PROXY": "http://proxy:8080"},
			expected: "http://proxy:8080",
		},
		{
			name:     "HARNESS prefixed env var set",
			key:      "http_proxy",
			envVars:  map[string]string{"HARNESS_HTTP_PROXY": "http://harness-proxy:8080"},
			expected: "http://harness-proxy:8080",
		},
		{
			name: "standard takes precedence over HARNESS",
			key:  "http_proxy",
			envVars: map[string]string{
				"HTTP_PROXY":         "http://standard:8080",
				"HARNESS_HTTP_PROXY": "http://harness:8080",
			},
			expected: "http://standard:8080",
		},
		{
			name: "lowercase takes precedence over uppercase",
			key:  "no_proxy",
			envVars: map[string]string{
				"no_proxy":        "localhost,127.0.0.1",
				"NO_PROXY":         "*.example.com",
				"HARNESS_NO_PROXY": "*.local",
			},
			expected: "localhost,127.0.0.1",
		},
		{
			name: "lowercase takes precedence over HARNESS",
			key:  "https_proxy",
			envVars: map[string]string{
				"https_proxy":        "https://standard:8080",
				"HARNESS_HTTPS_PROXY": "https://harness:8080",
			},
			expected: "https://standard:8080",
		},
		{
			name:     "no env var set",
			key:      "http_proxy",
			envVars:  map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			lowercaseKey := tt.key
			uppercaseKey := strings.ToUpper(tt.key)
			harnessKey := "HARNESS_" + strings.ToUpper(tt.key)

			os.Unsetenv(lowercaseKey)
			os.Unsetenv(uppercaseKey)
			os.Unsetenv(harnessKey)

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Execute and verify
			result := getProxyValue(tt.key)
			if result != tt.expected {
				t.Errorf("getProxyValue(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}
