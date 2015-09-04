package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/drone/drone-plugin-go/plugin"
)

type Docker struct {
	Storage  string   `json:"storage_driver"`
	Registry string   `json:"registry"`
	Insecure bool     `json:"insecure"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Auth     string   `json:"auth"`
	Repo     string   `json:"repo"`
	Tag      string   `json:"tag"`
	File     string   `json:"file"`
	Dns      []string `json:"dns"`
}

func main() {
	workspace := plugin.Workspace{}
	build := plugin.Build{}
	vargs := Docker{}

	plugin.Param("workspace", &workspace)
	plugin.Param("build", &build)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	// Set the storage driver
	if len(vargs.Storage) == 0 {
		vargs.Storage = "aufs"
	}

	stop := func() {
		cmd := exec.Command("start-stop-daemon", "--stop", "--pidfile", "/var/run/docker.pid")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		cmd.Run()
	}
	defer stop()

	// Starts the Docker daemon
	go func() {
		cmd := exec.Command("/bin/bash", "/bin/wrapdocker")
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		cmd.Run()

		args := []string{"daemon", "-s", vargs.Storage}

		if vargs.Insecure && len(vargs.Registry) != 0 {
			args = append(args, "--insecure-registry", vargs.Registry)
		}

		for _, value := range vargs.Dns {
			args = append(args, "--dns", value)
		}

		cmd = exec.Command("docker", args...)
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		trace(cmd)
		cmd.Run()
	}()

	// Sleep for a few seconds
	time.Sleep(5 * time.Second)

	// Set the Registry value
	if len(vargs.Registry) == 0 {
		vargs.Registry = "https://index.docker.io/v1/"
	} else {
		vargs.Repo = fmt.Sprintf("%s/%s", vargs.Registry, vargs.Repo)
	}
	// Set the Dockerfile path
	if len(vargs.File) == 0 {
		vargs.File = "."
	}
	// Set the Tag value
	if len(vargs.Tag) == 0 {
		vargs.Tag = "latest"
	}
	vargs.Repo = fmt.Sprintf("%s:%s", vargs.Repo, vargs.Tag)

	// Login to Docker
	if len(vargs.Username) != 0 {
		cmd := exec.Command("docker", "login", "-u", vargs.Username, "-p", vargs.Password, "-e", vargs.Email, vargs.Registry)
		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			stop()
			os.Exit(1)
		}
	} else {
		fmt.Printf("A username was not specified. Assuming anoynmous publishing.\n")
	}

	// Docker environment info
	cmd := exec.Command("docker", "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()
	cmd = exec.Command("docker", "info")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()

	// Build the container
	cmd = exec.Command("docker", "build", "--pull=true", "--rm=true", "-t", vargs.Repo, vargs.File)
	cmd.Dir = workspace.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	err := cmd.Run()
	if err != nil {
		stop()
		os.Exit(1)
	}

	// Push the container
	cmd = exec.Command("docker", "push", vargs.Repo)
	cmd.Dir = workspace.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	err = cmd.Run()
	if err != nil {
		stop()
		os.Exit(1)
	}
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}

// authorize is a helper function that authorizes the Docker client
// by manually creating the Docker authentication file.
func authorize(d *Docker) error {
	var path = "/root/.dockercfg" // TODO should probably use user.Home() for good measure
	var data = fmt.Sprintf(dockerconf, d.Registry, d.Auth, d.Email)
	return ioutil.WriteFile(path, []byte(data), 0644)
}

var dockerconf = `
{
	"%s": {
		"auth": "%s",
		"email": "%s"
	}
}
`
