package docker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
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
		DNSSearch     []string // Docker daemon dns search domain
		MTU           string   // Docker daemon mtu setting
		IPv6          bool     // Docker daemon IPv6 networking
		Experimental  bool     // Docker daemon enable experimental mode
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
		Remote      string   // Git remote URL
		Name        string   // Docker build using default named tag
		Dockerfile  string   // Docker build Dockerfile
		Context     string   // Docker build context
		Tags        []string // Docker build tags
		Args        []string // Docker build args
		ArgsEnv     []string // Docker build args from env
		Target      string   // Docker build target
		Squash      bool     // Docker build squash
		Pull        bool     // Docker build pull
		Compress    bool     // Docker build compress
		Repo        string   // Docker build repository
		LabelSchema []string // Label schema map
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login  Login  // Docker login configuration
		Build  Build  // Docker build configuration
		Daemon Daemon // Docker daemon configuration
		Dryrun bool   // Docker push is skipped
	}
)

// Exec executes the plugin step
func (p Plugin) Exec() error {
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

	if p.Build.Squash && !p.Daemon.Experimental {
		fmt.Println("Squash build flag is only available when Docker deamon is started with experimental flag. Ignoring...")
		p.Build.Squash = false
	}

	// add proxy build args
	addProxyBuildArgs(&p.Build)

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion()) // docker version
	cmds = append(cmds, commandInfo())    // docker info

	cmds = append(cmds, commandBuild(p.Build)) // docker build

	for _, tag := range p.Build.Tags {
		cmds = append(cmds, commandTag(p.Build, tag)) // docker tag

		if p.Dryrun == false {
			cmds = append(cmds, commandPush(p.Build, tag)) // docker push
		}
	}

	cmds = append(cmds, commandRmi(p.Build.Name)) // docker rmi
	cmds = append(cmds, commandPrune())           // docker system prune -f

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
const dockerdExe = "/usr/local/bin/dockerd"

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
	args := []string{
		"build",
		"--rm=true",
		"-f", build.Dockerfile,
		"-t", build.Name,
	}

	args = append(args, build.Context)
	if build.Squash {
		args = append(args, "--squash")
	}
	if build.Compress {
		args = append(args, "--compress")
	}
	if build.Pull {
		args = append(args, "--pull=true")
	}
	if build.Target != "" {
		args = append(args, "--target", build.Target)
	}
	for _, arg := range build.ArgsEnv {
		addProxyValue(&build, arg)
	}
	for _, arg := range build.Args {
		args = append(args, "--build-arg", arg)
	}

	labelSchema := []string{
		fmt.Sprintf("build-date=%s", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("vcs-ref=%s", build.Name),
		fmt.Sprintf("vcs-url=%s", build.Remote),
	}

	if len(build.LabelSchema) > 0 {
		labelSchema = append(labelSchema, build.LabelSchema...)
	}

	for _, label := range labelSchema {
		args = append(args, "--label", fmt.Sprintf("org.label-schema.%s", label))
	}

	return exec.Command(dockerExe, args...)
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
	args := []string{"-g", daemon.StoragePath}

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
	for _, dnsSearch := range daemon.DNSSearch {
		args = append(args, "--dns-search", dnsSearch)
	}
	if len(daemon.MTU) != 0 {
		args = append(args, "--mtu", daemon.MTU)
	}
	if daemon.Experimental {
		args = append(args, "--experimental")
	}
	return exec.Command(dockerdExe, args...)
}

func commandPrune() *exec.Cmd {
	return exec.Command(dockerExe, "system", "prune", "-f")
}

func commandRmi(tag string) *exec.Cmd {
	return exec.Command(dockerExe, "rmi", tag)
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
