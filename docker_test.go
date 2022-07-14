package docker

import (
	"os/exec"
	"reflect"
	"testing"
)

func Test_commandPush(t *testing.T) {
	type args struct {
		build Build
		tag   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "removes everything after last colon in CacheFrom[0]",
			args: args{
				build: Build{
					CacheBuilder: true,
					CacheFrom:    []string{"harbor.shipttech.com/buildcache/kubedashian-api:ncooke_ch339348_implement-caching-for-docker-in-docker-builds"},
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
		{
			name: "returns exact value plus tag in CacheRepo",
			args: args{
				build: Build{
					CacheBuilder: true,
					CacheRepo:    "harbor.shipttech.com/buildcache/kubedashian-api",
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
		{
			name: "handles no colon in CacheFrom[0]",
			args: args{
				build: Build{
					CacheBuilder: true,
					CacheFrom:    []string{"harbor.shipttech.com/buildcache/kubedashian-api"},
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
		{
			name: "returns Repo if CacheBuilder false",
			args: args{
				build: Build{
					CacheBuilder: false,
					CacheFrom:    []string{"no"},
					CacheRepo:    "not this",
					Repo:         "harbor.shipttech.com/buildcache/kubedashian-api",
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandPush(tt.args.build, tt.args.tag); !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("\ngot  = %v\nwant = %v", got.String(), tt.want)
			}
		})
	}
}

func TestCommandBuild(t *testing.T) {
	tcs := []struct {
		name  string
		build Build
		want  *exec.Cmd
	}{
		{
			name: "secret from env var",
			build: Build{
				Name:       "plugins/drone-docker:latest",
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
				"plugins/drone-docker:latest",
				".",
				"--secret id=foo_secret,env=FOO_SECRET_ENV_VAR",
			),
		},
		{
			name: "secret from file",
			build: Build{
				Name:       "plugins/drone-docker:latest",
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
				"plugins/drone-docker:latest",
				".",
				"--secret id=foo_secret,src=/path/to/foo_secret",
			),
		},
		{
			name: "multiple mixed secrets",
			build: Build{
				Name:       "plugins/drone-docker:latest",
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
				"plugins/drone-docker:latest",
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
				"plugins/drone-docker:latest",
				".",
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
