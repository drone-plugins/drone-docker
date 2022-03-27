package docker

import (
	"fmt"
	"github.com/drone-plugins/drone-docker/utils"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		Config   string // Docker Auth Config
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
		CacheFrom   []string // Docker build cache-from
		Compress    bool     // Docker build compress
		Repo        string   // Docker build repository
		LabelSchema []string // label-schema Label map
		AutoLabel   bool     // auto-label bool
		Labels      []string // Label map
		Link        string   // Git repo link
		NoCache     bool     // Docker build no-cache
		Secret      string   // secret keypair
		SecretEnvs  []string // Docker build secrets with env var as source
		SecretFiles []string // Docker build secrets with file as source
		AddHost     []string // Docker build add-host
		Quiet       bool     // Docker build quiet
	}

	Buildx struct {
		ConfigFile string
		Driver     string
		DriverOpts []string
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login    Login      // Docker login configuration
		Build    Build      // Docker build configuration
		Buildx   Buildx     // Docker buildx configuration
		Daemon   Daemon     // Docker daemon configuration
		Dryrun   bool       // Docker push is skipped
		Cleanup  bool       // Docker purge is enabled
		CardPath string     // Card path to write file to
		Executor utils.Exec // Wrapper to allow mocking os.exec calls
	}

	Card []struct {
		ID             string        `json:"Id"`
		RepoTags       []string      `json:"RepoTags"`
		ParsedRepoTags []TagStruct   `json:"ParsedRepoTags"`
		RepoDigests    []interface{} `json:"RepoDigests"`
		Parent         string        `json:"Parent"`
		Comment        string        `json:"Comment"`
		Created        time.Time     `json:"Created"`
		Container      string        `json:"Container"`
		DockerVersion  string        `json:"DockerVersion"`
		Author         string        `json:"Author"`
		Architecture   string        `json:"Architecture"`
		Os             string        `json:"Os"`
		Size           int           `json:"Size"`
		VirtualSize    int           `json:"VirtualSize"`
		Metadata       struct {
			LastTagTime time.Time `json:"LastTagTime"`
		} `json:"Metadata"`
		SizeString        string
		VirtualSizeString string
		Time              string
		URL               string `json:"URL"`
	}
	TagStruct struct {
		Tag string `json:"Tag"`
	}
)

// Exec executes the plugin step
func (p Plugin) Exec() (err error) {
	// start the Docker daemon server
	if !p.Daemon.Disabled {
		p.startDaemon()
	}

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; ; i++ {
		if err := p.Executor.Command(dockerExe, "info").Run(); err == nil {
			break
		}
		if i == 15 {
			fmt.Println("Unable to reach Docker Daemon after 15 attempts.")
			break
		}
		time.Sleep(time.Second * 1)
	}

	if err := p.printDockerVersion(); err != nil {
		return err
	}

	if err := p.printDockerInfo(); err != nil {
		return err
	}

	// for debugging purposes, log the type of authentication
	// credentials that have been provided.
	switch {
	case p.Login.Password != "" && p.Login.Config != "":
		fmt.Println("Detected registry credentials and registry credentials file")
	case p.Login.Password != "":
		fmt.Println("Detected registry credentials")
	case p.Login.Config != "":
		fmt.Println("Detected registry credentials file")
	default:
		fmt.Println("Registry credentials or Docker config not provided. Guest mode enabled.")
	}

	// create Auth Config File
	if p.Login.Config != "" {
		os.MkdirAll(dockerHome, 0600)

		path := filepath.Join(dockerHome, "config.json")
		err := ioutil.WriteFile(path, []byte(p.Login.Config), 0600)
		if err != nil {
			return fmt.Errorf("Error writing config.json: %s", err)
		}
	}

	// login to the Docker registry
	if p.Login.Password != "" {
		cmd := commandLogin(p.Login)
		raw, err := cmd.CombinedOutput()
		if err != nil {
			out := string(raw)
			out = strings.Replace(out, "WARNING! Using --password via the CLI is insecure. Use --password-stdin.", "", -1)
			fmt.Println(out)
			return fmt.Errorf("Error authenticating: exit status 1")
		}
	}

	if p.Build.Squash && !p.Daemon.Experimental {
		fmt.Println("Squash build flag is only available when Docker deamon is started with experimental flag. Ignoring...")
		p.Build.Squash = false
	}

	// add proxy build args
	addProxyBuildArgs(&p.Build)

	builderName := "default"
	if p.Buildx.Driver != "docker" {
		builderName = "builder-" + uuid.NewString()
	}

	if err := p.createBuildxBuilder(builderName); err != nil {
		return err
	}

	defer func(name string, buildx Buildx) {
		if tempErr := p.removeBuildxBuilder(name); tempErr != nil {
			err = tempErr
		}
	}(builderName, p.Buildx)

	if err := p.inspectBuildxBuilder(builderName); err != nil {
		return err
	}

	for _, img := range p.Build.CacheFrom {
		if err := p.pullDockerImage(img); err != nil {
			fmt.Printf("Could not pull cache-from image %s. Ignoring...\n", img)
		}
	}

	if err := p.buildDockerImage(); err != nil {
		return err
	}

	if err := p.printAdaptiveCard(); err != nil {
		fmt.Printf("Could not create adaptive card. %s\n", err)
	}

	return p.pruneDocker()
}

// helper function to create the docker login command.
func commandLogin(login Login) *exec.Cmd {
	if login.Email != "" {
		return commandLoginEmail(login)
	}
	return exec.Command(
		dockerExe,
		"login",
		"-u", login.Username,
		"-p", login.Password,
		login.Registry,
	)
}

func commandLoginEmail(login Login) *exec.Cmd {
	return exec.Command(
		dockerExe,
		"login",
		"-u", login.Username,
		"-p", login.Password,
		"-e", login.Email,
		login.Registry,
	)
}

func (p *Plugin) printDockerVersion() error {
	cmd := p.Executor.Command(dockerExe, "version")
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get docker version: %w", err)
	}

	return nil
}

func (p *Plugin) printDockerInfo() error {
	cmd := p.Executor.Command(dockerExe, "info")
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get docker info: %w", err)
	}

	return nil
}

func (p *Plugin) pruneDocker() error {
	if p.Cleanup {
		cmd := p.Executor.Command(dockerExe, "system", "prune", "-f")
		cmd.SetStdout(os.Stdout)
		cmd.SetStderr(os.Stderr)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to prune unused docker containers, networks and images: %w", err)
		}
	}

	return nil
}

func (p *Plugin) pullDockerImage(img string) error {
	cmd := p.Executor.Command(dockerExe, "pull", img)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull docker image: %w", err)
	}

	return nil
}

func (p *Plugin) buildDockerImage() error {
	args := []string{
		"buildx",
		"build",
		"--rm",
		"--file",
		p.Build.Dockerfile,
		"--tag",
		p.Build.Name,
	}

	if p.Build.Squash {
		args = append(args, "--squash")
	}
	if p.Build.Compress {
		args = append(args, "--compress")
	}
	if p.Build.Pull {
		args = append(args, "--pull=true")
	}
	if p.Build.NoCache {
		args = append(args, "--no-cache")
	}
	for _, arg := range p.Build.CacheFrom {
		args = append(args, "--cache-from", arg)
	}
	for _, arg := range p.Build.ArgsEnv {
		addProxyValue(&p.Build, arg)
	}
	for _, arg := range p.Build.Args {
		args = append(args, "--build-arg", arg)
	}
	for _, host := range p.Build.AddHost {
		args = append(args, "--add-host", host)
	}
	if p.Build.Secret != "" {
		args = append(args, "--secret", p.Build.Secret)
	}
	for _, secret := range p.Build.SecretEnvs {
		if arg, err := getSecretStringCmdArg(secret); err == nil {
			args = append(args, "--secret", arg)
		}
	}
	for _, secret := range p.Build.SecretFiles {
		if arg, err := getSecretFileCmdArg(secret); err == nil {
			args = append(args, "--secret", arg)
		}
	}
	if p.Build.Target != "" {
		args = append(args, "--target", p.Build.Target)
	}
	if p.Build.Quiet {
		args = append(args, "--quiet")
	}

	if p.Build.AutoLabel {
		labelSchema := []string{
			fmt.Sprintf("created=%s", time.Now().Format(time.RFC3339)),
			fmt.Sprintf("revision=%s", p.Build.Name),
			fmt.Sprintf("source=%s", p.Build.Remote),
			fmt.Sprintf("url=%s", p.Build.Link),
		}
		labelPrefix := "org.opencontainers.image"

		if len(p.Build.LabelSchema) > 0 {
			labelSchema = append(labelSchema, p.Build.LabelSchema...)
		}

		for _, label := range labelSchema {
			args = append(args, "--label", fmt.Sprintf("%s.%s", labelPrefix, label))
		}
	}

	if len(p.Build.Labels) > 0 {
		for _, label := range p.Build.Labels {
			args = append(args, "--label", label)
		}
	}

	for _, tag := range p.Build.Tags {
		args = append(args, "--tag", fmt.Sprintf("%s:%s", p.Build.Repo, tag))
	}

	if !p.Dryrun {
		args = append(args, "--push")
	}

	args = append(args, p.Build.Context)

	cmd := p.Executor.Command(dockerExe, args...)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	return nil
}

func getSecretStringCmdArg(kvp string) (string, error) {
	return getSecretCmdArg(kvp, false)
}

func getSecretFileCmdArg(kvp string) (string, error) {
	return getSecretCmdArg(kvp, true)
}

func getSecretCmdArg(kvp string, file bool) (string, error) {
	delimIndex := strings.IndexByte(kvp, '=')
	if delimIndex == -1 {
		return "", fmt.Errorf("%s is not a valid secret", kvp)
	}

	key := kvp[:delimIndex]
	value := kvp[delimIndex+1:]

	if key == "" || value == "" {
		return "", fmt.Errorf("%s is not a valid secret", kvp)
	}

	if file {
		return fmt.Sprintf("id=%s,src=%s", key, value), nil
	}

	return fmt.Sprintf("id=%s,env=%s", key, value), nil
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

// helper function to create the docker daemon command.
func commandDaemon(daemon Daemon) *exec.Cmd {
	args := []string{
		"--data-root", daemon.StoragePath,
		"--host=unix:///var/run/docker.sock",
	}

	if _, err := os.Stat("/etc/docker/default.json"); err == nil {
		args = append(args, "--seccomp-profile=/etc/docker/default.json")
	}

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

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Printf("+ %s\n", strings.Join(cmd.Args, " "))
}

func GetDroneDockerExecCmd() string {
	if runtime.GOOS == "windows" {
		return "C:/bin/drone-docker.exe"
	}

	return "drone-docker"
}
