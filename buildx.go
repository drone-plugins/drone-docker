package docker

import (
	"fmt"
	"os"
)

func (p *Plugin) createBuildxBuilder(name string) error {
	if p.Buildx.Driver != "docker" {
		args := []string{"buildx", "create", "--use", "--name", name, "--driver", p.Buildx.Driver}

		if p.Buildx.ConfigFile != "" {
			args = append(args, "--config", p.Buildx.ConfigFile)
		}

		for _, opt := range p.Buildx.DriverOpts {
			args = append(args, "--driver-opt", opt)
		}

		cmd := p.Executor.Command(dockerExe, args...)
		cmd.SetStdout(os.Stdout)
		cmd.SetStderr(os.Stderr)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error booting the docker buildx builder: %v", err)
		}
	}

	return nil
}

func (p *Plugin) removeBuildxBuilder(name string) error {
	if p.Buildx.Driver != "docker" {
		cmd := p.Executor.Command(dockerExe, "buildx", "rm", name)
		cmd.SetStdout(os.Stdout)
		cmd.SetStderr(os.Stderr)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to remove the buildx builder: %v", err)
		}
	}

	return nil
}

func (p *Plugin) inspectBuildxBuilder(name string) error {
	cmd := p.Executor.Command(dockerExe, "buildx", "inspect", "--bootstrap", "--builder", name)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to inspect the buildx builder: %v", err)
	}

	return nil
}
