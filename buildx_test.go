package docker

import (
	"github.com/stretchr/testify/require"
	utilexec "k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
	"testing"
)

func TestPlugin_CreateBuildxBuilderDoesNothingIfDriverIsDocker(t *testing.T) {
	fexec := fakeexec.FakeExec{}
	p := Plugin{
		Executor: &fexec,
		Buildx: Buildx{
			Driver: "docker",
		},
	}

	require.NoError(t, p.createBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e"))
	require.Equal(t, 0, fexec.CommandCalls)
}

func TestPlugin_CreateBuildxBuilderSucceedsIfDriverIsDockerContainer(t *testing.T) {
	tcs := []struct {
		name       string
		buildxConf Buildx
		arguments  []string
	}{
		{
			name: "Default config options",
			buildxConf: Buildx{
				Driver: "docker-container",
			},
			arguments: []string{
				dockerExe,
				"buildx",
				"create",
				"--use",
				"--name",
				"builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e",
				"--driver",
				"docker-container",
			},
		},
		{
			name: "Config file specified",
			buildxConf: Buildx{
				Driver:     "docker-container",
				ConfigFile: "/path/to/config",
			},
			arguments: []string{
				dockerExe,
				"buildx",
				"create",
				"--use",
				"--name", "builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e",
				"--driver",
				"docker-container",
				"--config",
				"/path/to/config",
			},
		},
		{
			name: "Driver options specified",
			buildxConf: Buildx{
				Driver:     "docker-container",
				DriverOpts: []string{"foo=bar", "bar=foo"},
			},
			arguments: []string{
				dockerExe,
				"buildx",
				"create",
				"--use",
				"--name", "builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e",
				"--driver",
				"docker-container",
				"--driver-opt",
				"foo=bar",
				"--driver-opt",
				"bar=foo",
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
				Executor: &fakeexec.FakeExec{
					CommandScript: []fakeexec.FakeCommandAction{
						func(cmd string, args ...string) utilexec.Cmd {
							return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
						},
					},
				},
				Buildx: tc.buildxConf,
			}

			require.NoError(t, p.createBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e"))
			require.Equal(t, 1, fcmd.RunCalls)
			require.Equal(t, tc.arguments, fcmd.Argv)
		})
	}
}

func TestPlugin_CreateBuildxBuilderReturnsErrorIfCreationFails(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) {
				return nil, nil, &fakeexec.FakeExitError{Status: 1}
			},
		},
	}
	p := Plugin{
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		},
		Buildx: Buildx{
			Driver: "docker-container",
		},
	}

	err := p.createBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e")

	require.Error(t, err)
	require.ErrorContains(t, err, "error booting the docker buildx builder: exit 1")
	require.Equal(t, 1, fcmd.RunCalls)
}

func TestPlugin_RemoveBuildxBuilderDoesNothingIfDriverIsDocker(t *testing.T) {
	fexec := fakeexec.FakeExec{}
	p := Plugin{
		Executor: &fexec,
		Buildx: Buildx{
			Driver: "docker",
		},
	}

	require.NoError(t, p.removeBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e"))
	require.Equal(t, 0, fexec.CommandCalls)
}

func TestPlugin_RemoveBuildxBuilderSucceedsIfDriverIsDockerContainer(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) {
				return nil, nil, nil
			},
		},
	}
	p := Plugin{
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		},
		Buildx: Buildx{
			Driver: "docker-container",
		},
	}

	require.NoError(t, p.removeBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e"))
	require.Equal(t, 1, fcmd.RunCalls)
}

func TestPlugin_RemoveBuildxBuilderReturnsErrorIfRemovalFails(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) {
				return nil, nil, &fakeexec.FakeExitError{Status: 1}
			},
		},
	}
	p := Plugin{
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		},
		Buildx: Buildx{
			Driver: "docker-container",
		},
	}

	err := p.removeBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e")

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to remove the buildx builder: exit 1")
	require.Equal(t, 1, fcmd.RunCalls)
}

func TestPlugin_InspectBuildxBuilder(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) {
				return nil, nil, nil
			},
		},
	}
	p := Plugin{
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		},
	}

	require.NoError(t, p.inspectBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e"))
	require.Equal(t, 1, fcmd.RunCalls)
}

func TestPlugin_InspectBuildxBuilderReturnsErrorIfInspectionFails(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		RunScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) {
				return nil, nil, &fakeexec.FakeExitError{Status: 1}
			},
		},
	}
	p := Plugin{
		Executor: &fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilexec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		},
	}

	err := p.inspectBuildxBuilder("builder-84a44a22-398c-45fe-9eaf-4ceac27ddc6e")

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to inspect the buildx builder: exit 1")
	require.Equal(t, 1, fcmd.RunCalls)
}
