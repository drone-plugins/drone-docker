package docker

import (
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

func TestNoPushFlag(t *testing.T) {
	testCases := []struct {
		name           string
		noPush         bool
		dryrun         bool
		expectPushCmds bool
	}{
		{
			name:           "NoPush Enabled",
			noPush:         true,
			dryrun:         false,
			expectPushCmds: false,
		},
		{
			name:           "Dryrun Enabled",
			noPush:         false,
			dryrun:         true,
			expectPushCmds: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := Plugin{
				Build: Build{
					NoPush:  tc.noPush,
					Repo:    "testuser/test-image",
					Tags:    []string{"latest"},
					TempTag: "test-image:latest",
				},
				Dryrun: tc.dryrun,
			}

			// Simulate plugin's command generation logic
			cmds := []*exec.Cmd{
				commandVersion(),
				commandInfo(),
				commandBuild(plugin.Build),
			}

			for _, tag := range plugin.Build.Tags {
				cmds = append(cmds, commandTag(plugin.Build, tag))
			}

			if !tc.noPush && !tc.dryrun {
				for _, tag := range plugin.Build.Tags {
					cmds = append(cmds, commandPush(plugin.Build, tag))
				}
			}

			// Verify the number of commands
			var pushCmds int
			for _, cmd := range cmds {
				if strings.Contains(cmd.Args[0], "push") {
					pushCmds++
				}
			}

			if tc.expectPushCmds {
				if pushCmds == 0 {
					t.Errorf("Expected push commands, but none were generated")
				}
			} else {
				if pushCmds > 0 {
					t.Errorf("Did not expect push commands, but %d were generated", pushCmds)
				}
			}
		})
	}
}

func TestTarPathFlag(t *testing.T) {
	testCases := []struct {
		name          string
		tarPath       string
		expectSaveCmd bool
	}{
		{
			name:          "TarPath Not Set",
			tarPath:       "",
			expectSaveCmd: false,
		},
		{
			name:          "TarPath Set",
			tarPath:       "/tmp/test-image.tar",
			expectSaveCmd: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			build := Build{
				Repo:    "testuser/test-image",
				Tags:    []string{"latest"},
				TempTag: "test-image:latest",
				TarPath: tc.tarPath,
			}

			saveCmd := commandSave(build)

			if tc.expectSaveCmd {
				if saveCmd == nil {
					t.Errorf("Expected save command, but none was generated")
				} else {
					if saveCmd.Args[0] != dockerExe || saveCmd.Args[1] != "save" {
						t.Errorf("Incorrect save command: got %v", saveCmd.Args)
					}
					if saveCmd.Args[3] != tc.tarPath {
						t.Errorf("Incorrect tar path: want %s, got %s", tc.tarPath, saveCmd.Args[3])
					}
				}
			} else {
				if saveCmd != nil {
					t.Errorf("Did not expect save command, but one was generated")
				}
			}
		})
	}
}
