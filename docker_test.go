package docker

import (
	"github.com/stretchr/testify/require"
	utilexec "k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
	"testing"
)

func TestPlugin_Exec(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, nil },
		},
	}

	p := Plugin{
		Cleanup: true,
		Buildx: Buildx{
			Driver: "docker",
		},
		Build: Build{
			Name:       "plugins/drone-docker:latest",
			Dockerfile: "Dockerfile",
			Context:    ".",
		},
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
			},
		},
	}

	expectedCmds := [][]string{
		{dockerExe, "info"},
		{dockerExe, "version"},
		{dockerExe, "info"},
		{dockerExe, "buildx", "inspect", "--bootstrap", "--builder", "default"},
		{dockerExe, "buildx", "build", "--rm", "--file", "Dockerfile", "--tag", "plugins/drone-docker:latest", "--push", "."},
		{dockerExe, "system", "prune", "-f"},
	}

	require.NoError(t, p.Exec())
	require.Equal(t, 6, fcmd.RunCalls)
	require.Equal(t, expectedCmds, fcmd.RunLog)
}

func TestPlugin_ExecReturnsErrorIfDockerVersionCannotBeRetrieved(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, &fakeexec.FakeExitError{Status: 1} },
		},
	}

	p := Plugin{
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
				func(cmd string, args ...string) utilexec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		},
	}

	require.Error(t, p.Exec())
	require.Equal(t, 2, fcmd.RunCalls)
	require.Equal(t, []string{dockerExe, "version"}, fcmd.RunLog[1])
}

func TestPlugin_ExecReturnsErrorIfDockerInfoCannotBeRetrieved(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, nil },
			func() ([]byte, []byte, error) { return nil, nil, &fakeexec.FakeExitError{Status: 1} },
		},
	}

	p := Plugin{
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
				func(cmd string, args ...string) utilexec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
			},
		},
	}

	require.Error(t, p.Exec())
	require.Equal(t, 3, fcmd.RunCalls)
	require.Equal(t, []string{dockerExe, "info"}, fcmd.RunLog[2])
}

func TestPlugin_BuildDockerImage(t *testing.T) {
	tcs := []struct {
		name      string
		build     Build
		dryRun    bool
		arguments []string
	}{
		{
			name: "image is pushed to registry when dry-run mode is not enabled",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				Dockerfile: "Dockerfile",
				Context:    ".",
			},
			dryRun: false,
			arguments: []string{
				dockerExe,
				"buildx",
				"build",
				"--rm",
				"--file",
				"Dockerfile",
				"--tag",
				"plugins/drone-docker:latest",
				"--push",
				".",
			},
		},
		{
			name: "the image is tagged according to the configuration",
			build: Build{
				Name:       "plugins/drone-docker:latest",
				Dockerfile: "Dockerfile",
				Context:    ".",
				Repo:       "plugins/drone-docker",
				Tags: []string{
					"v0",
					"v0.0.0",
				},
			},
			dryRun: true,
			arguments: []string{
				dockerExe,
				"buildx",
				"build",
				"--rm",
				"--file",
				"Dockerfile",
				"--tag",
				"plugins/drone-docker:latest",
				"--tag",
				"plugins/drone-docker:v0",
				"--tag",
				"plugins/drone-docker:v0.0.0",
				".",
			},
		},
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
			arguments: []string{
				dockerExe,
				"buildx",
				"build",
				"--rm",
				"--file",
				"Dockerfile",
				"--tag",
				"plugins/drone-docker:latest",
				"--secret",
				"id=foo_secret,env=FOO_SECRET_ENV_VAR",
				"--push",
				".",
			},
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
			arguments: []string{
				dockerExe,
				"buildx",
				"build",
				"--rm",
				"--file",
				"Dockerfile",
				"--tag",
				"plugins/drone-docker:latest",
				"--secret",
				"id=foo_secret,src=/path/to/foo_secret",
				"--push",
				".",
			},
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
			arguments: []string{
				dockerExe,
				"buildx",
				"build",
				"--rm",
				"--file",
				"Dockerfile",
				"--tag",
				"plugins/drone-docker:latest",
				"--secret",
				"id=foo_secret,env=FOO_SECRET_ENV_VAR",
				"--secret",
				"id=bar_secret,env=BAR_SECRET_ENV_VAR",
				"--secret",
				"id=foo_secret,src=/path/to/foo_secret",
				"--secret",
				"id=bar_secret,src=/path/to/bar_secret",
				"--push",
				".",
			},
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
			arguments: []string{
				dockerExe,
				"buildx",
				"build",
				"--rm",
				"--file",
				"Dockerfile",
				"--tag",
				"plugins/drone-docker:latest",
				"--push",
				".",
			},
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			fcmd := fakeexec.FakeCmd{
				RunScript: []fakeexec.FakeAction{
					func() ([]byte, []byte, error) {
						return nil, nil, nil
					},
				},
			}
			p := Plugin{
				Build:  tc.build,
				Dryrun: tc.dryRun,
				Executor: &fakeexec.FakeExec{
					CommandScript: []fakeexec.FakeCommandAction{
						func(cmd string, args ...string) utilexec.Cmd {
							return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
						},
					},
				},
			}

			require.NoError(t, p.buildDockerImage())
			require.Equal(t, 1, fcmd.RunCalls)
			require.Equal(t, tc.arguments, fcmd.Argv)
		})
	}
}
