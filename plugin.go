package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const (
	// default docker registry
	defaultRegistry = "https://index.docker.io/v1/"
)

type (
	// Daemon defines Docker daemon parameters.
	Daemon struct {
		Registry      string   // Docker registry
		Mirror        string   // Docker registry mirror
		Insecure      bool     // Docker daemon enable insecure registries
		StorageDriver string   // Docker daemon storage driver
		StoragePath   string   // Docker daemon storage path
		Disabled      bool     // DOcker daemon is disabled (already running)
		Debug         bool     // Docker daemon started in debug mode
		Bip           string   // Docker daemon network bridge IP address
		DNS           []string // Docker daemon dns server
		MTU           string   // Docker daemon mtu setting
		IPv6          bool     // Docker daemon IPv6 networking
	}

	// Login defines Docker login parameters.
	Login struct {
		Registry string // Docker registry address
		Username string // Docker registry username
		Password string // Docker registry password
		Email    string // Docker registry email
	}

	// Build defines Docker build parameters.
	Build struct {
		Name       string   // Docker build using default named tag
		Dockerfile string   // Docker build Dockerfile
		Context    string   // Docker build context
		Tags       []string // Docker build tags
		Args       []string // Docker build args
		Repo       string   // Docker build repository
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login  Login  // Docker login configuration
		Build  Build  // Docker build configuration
		Daemon Daemon // Docker daemon configuration
		Dryrun bool   // Docker push is skipped
	}

	// TagsFile defines slice of tags from .droneTags.yml
	TagsFile struct {
		Tags []string `yaml:"tags"`
	}
)

// Exec executes the plugin step
func (p Plugin) Exec() error {

	// TODO execute code remove dangling images
	// this is problematic because we are running docker in scratch which does
	// not have bash, so we need to hack something together
	// docker images --quiet --filter=dangling=true | xargs --no-run-if-empty docker rmi

	/*
		cmd = exec.Command("docker", "images", "-q", "-f", "dangling=true")
		cmd = exec.Command("docker", append([]string{"rmi"}, images...)...)
	*/

	// start the Docker daemon server
	if !p.Daemon.Disabled {
		cmd := commandDaemon(p.Daemon)
		if p.Daemon.Debug {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = ioutil.Discard
			cmd.Stderr = ioutil.Discard
		}
		go func() {
			trace(cmd)
			cmd.Run()
		}()
	}

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; i < 15; i++ {
		cmd := commandInfo()
		err := cmd.Run()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}

	// login to the Docker registry
	if p.Login.Password != "" {
		cmd := commandLogin(p.Login)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Error authenticating: %s", err)
		}
	} else {
		fmt.Println("Registry credentials not provided. Guest mode enabled.")
	}

	// add proxy build args
	addProxyBuildArgs(&p.Build)

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion())      // docker version
	cmds = append(cmds, commandInfo())         // docker info
	cmds = append(cmds, commandBuild(p.Build)) // docker build

	// Override Tags if .droneTags.yml exists
	droneTags, err := getDroneTags(p.Build.Context)
	if err != nil {
		return err
	}
	if droneTags != nil {
		p.Build.Tags = droneTags
	}

	for _, tag := range p.Build.Tags {
		cmds = append(cmds, commandTag(p.Build, tag)) // docker tag

		if p.Dryrun == false {
			cmds = append(cmds, commandPush(p.Build, tag)) // docker push
		}
	}

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

const dockerExe = "/usr/local/bin/docker"

// helper function to create the docker login command.
func commandLogin(login Login) *exec.Cmd {
	if login.Email != "" {
		return commandLoginEmail(login)
	}
	return exec.Command(
		dockerExe, "login",
		"-u", login.Username,
		"-p", login.Password,
		login.Registry,
	)
}

func commandLoginEmail(login Login) *exec.Cmd {
	return exec.Command(
		dockerExe, "login",
		"-u", login.Username,
		"-p", login.Password,
		"-e", login.Email,
		login.Registry,
	)
}

// helper function to create the docker info command.
func commandVersion() *exec.Cmd {
	return exec.Command(dockerExe, "version")
}

// helper function to create the docker info command.
func commandInfo() *exec.Cmd {
	return exec.Command(dockerExe, "info")
}

// helper function to create the docker build command.
func commandBuild(build Build) *exec.Cmd {
	cmd := exec.Command(
		dockerExe, "build",
		"--pull=true",
		"--rm=true",
		"-f", build.Dockerfile,
		"-t", build.Name,
	)

	for _, arg := range build.Args {
		cmd.Args = append(cmd.Args, "--build-arg", arg)
	}
	cmd.Args = append(cmd.Args, build.Context)
	return cmd
}

// helper function to add proxy values from the environment
func addProxyBuildArgs(build *Build) {
	addProxyValue(build, "http_proxy")
	addProxyValue(build, "https_proxy")
	addProxyValue(build, "no_proxy")
}

// helper function to add the upper and lower case version of a proxy value.
func addProxyValue(build *Build, key string) {
	value := getProxyValue(key)

	if len(value) > 0 && !hasProxyBuildArg(build, key) {
		build.Args = append(build.Args, fmt.Sprintf("%s=%s", key, value))
		build.Args = append(build.Args, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}
}

// helper function to get a proxy value from the environment.
//
// assumes that the upper and lower case versions of are the same.
func getProxyValue(key string) string {
	value := os.Getenv(key)

	if len(value) > 0 {
		return value
	}

	return os.Getenv(strings.ToUpper(key))
}

// helper function that looks to see if a proxy value was set in the build args.
func hasProxyBuildArg(build *Build, key string) bool {
	keyUpper := strings.ToUpper(key)

	for _, s := range build.Args {
		if strings.HasPrefix(s, key) || strings.HasPrefix(s, keyUpper) {
			return true
		}
	}

	return false
}

// helper function to create the docker tag command.
func commandTag(build Build, tag string) *exec.Cmd {
	var (
		source = build.Name
		target = fmt.Sprintf("%s:%s", build.Repo, tag)
	)
	return exec.Command(
		dockerExe, "tag", source, target,
	)
}

// helper function to create the docker push command.
func commandPush(build Build, tag string) *exec.Cmd {
	target := fmt.Sprintf("%s:%s", build.Repo, tag)
	return exec.Command(dockerExe, "push", target)
}

// helper function to create the docker daemon command.
func commandDaemon(daemon Daemon) *exec.Cmd {
	args := []string{"daemon", "-g", daemon.StoragePath}

	if daemon.StorageDriver != "" {
		args = append(args, "-s", daemon.StorageDriver)
	}
	if daemon.Insecure && daemon.Registry != "" {
		args = append(args, "--insecure-registry", daemon.Registry)
	}
	if daemon.IPv6 {
		args = append(args, "--ipv6")
	}
	if len(daemon.Mirror) != 0 {
		args = append(args, "--registry-mirror", daemon.Mirror)
	}
	if len(daemon.Bip) != 0 {
		args = append(args, "--bip", daemon.Bip)
	}
	for _, dns := range daemon.DNS {
		args = append(args, "--dns", dns)
	}
	if len(daemon.MTU) != 0 {
		args = append(args, "--mtu", daemon.MTU)
	}
	return exec.Command(dockerExe, args...)
}

// conditionally read in .droneTags.yml and return resulting data structure
func getDroneTags(context string) ([]string, error) {
	fmt.Println("Looking For .droneTags.yml")
	var path = filepath.Join(context, ".droneTags.yml")
	_, err := os.Stat(path)
	if err != nil {
		// droneTags.yml doesn't exist, this is OK
		return nil, nil
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error Reading File %s, %s", path, err)
	}

	fmt.Println("Found .droneTags.yml")
	var droneTags TagsFile
	err = yaml.Unmarshal(file, &droneTags)
	if err != nil {
		return nil, fmt.Errorf("Could not parse .droneTags.yml: %s", err)
	}

	fmt.Println("Tag generated from .droneTags.yml:")
	for _, tag := range droneTags.Tags {
		fmt.Println("  ", tag)
	}
	return droneTags.Tags, nil
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
