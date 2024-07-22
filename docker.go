package docker

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/drone-plugins/drone-docker/internal/docker"
	"github.com/drone-plugins/drone-plugin-lib/drone"
)

type (
	// Daemon defines Docker daemon parameters.
	Daemon struct {
		Registry      string             // Docker registry
		Mirror        string             // Docker registry mirror
		Insecure      bool               // Docker daemon enable insecure registries
		StorageDriver string             // Docker daemon storage driver
		StoragePath   string             // Docker daemon storage path
		Disabled      bool               // DOcker daemon is disabled (already running)
		Debug         bool               // Docker daemon started in debug mode
		Bip           string             // Docker daemon network bridge IP address
		DNS           []string           // Docker daemon dns server
		DNSSearch     []string           // Docker daemon dns search domain
		MTU           string             // Docker daemon mtu setting
		IPv6          bool               // Docker daemon IPv6 networking
		Experimental  bool               // Docker daemon enable experimental mode
		RegistryType  drone.RegistryType // Docker registry type
	}

	// Login defines Docker login parameters.
	Login struct {
		Registry    string // Docker registry address
		Username    string // Docker registry username
		Password    string // Docker registry password
		Email       string // Docker registry email
		Config      string // Docker Auth Config
		AccessToken string // External Access Token
	}

	// Build defines Docker build parameters.
	Build struct {
		Remote      string   // Git remote URL
		Name        string   // Docker build using default named tag
		TempTag     string   // Temporary tag used during docker build
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
		Platform    string   // Docker build platform
		SSHAgentKey string   // Docker build ssh agent key
		SSHKeyPath  string   // Docker build ssh key path
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login             Login  // Docker login configuration
		Build             Build  // Docker build configuration
		Daemon            Daemon // Docker daemon configuration
		Dryrun            bool   // Docker push is skipped
		Cleanup           bool   // Docker purge is enabled
		CardPath          string // Card path to write file to
		ArtifactFile      string // Artifact path to write file to
		BaseImageRegistry string // Docker registry to pull base image
		BaseImageUsername string // Docker registry username to pull base image
		BaseImagePassword string // Docker registry password to pull base image
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
func (p Plugin) Exec() error {
	// start the Docker daemon server
	if !p.Daemon.Disabled {
		p.startDaemon()
	}

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; ; i++ {
		cmd := commandInfo()
		err := cmd.Run()
		if err == nil {
			break
		}
		if i == 15 {
			fmt.Println("Unable to reach Docker Daemon after 15 attempts.")
			break
		}
		time.Sleep(time.Second * 1)
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
	case p.Login.AccessToken != "":
		fmt.Println("Detected access token")
	default:
		fmt.Println("Registry credentials or Docker config not provided. Guest mode enabled.")
	}

	// create Auth Config File
	if p.Login.Config != "" {
		os.MkdirAll(dockerHome, 0600)

		path := filepath.Join(dockerHome, "config.json")
		err := os.WriteFile(path, []byte(p.Login.Config), 0600)
		if err != nil {
			return fmt.Errorf("Error writing config.json: %s", err)
		}
	}

	// instead of writing to config file directly, using docker's login func
	// is better to integrate with various credential helpers,
	//	it also handles different registry specific logic in a better way,
	//	as opposed to config write where different registries need to be addressed differently.
	//	It handles any changes in the authentication process across different Docker versions.

	if p.BaseImageRegistry != "" {
		if p.BaseImageUsername == "" {
			return fmt.Errorf("Username cannot be empty. The base image connector requires authenticated access. Please either use an authenticated connector, or remove the base image connector.")
		}
		if p.BaseImagePassword == "" {
			return fmt.Errorf("Password cannot be empty. The base image connector requires authenticated access. Please either use an authenticated connector, or remove the base image connector.")
		}
		var baseConnectorLogin Login
		baseConnectorLogin.Registry = p.BaseImageRegistry
		baseConnectorLogin.Username = p.BaseImageUsername
		baseConnectorLogin.Password = p.BaseImagePassword

		cmd := commandLogin(baseConnectorLogin)

		raw, err := cmd.CombinedOutput()
		if err != nil {
			out := string(raw)
			out = strings.Replace(out, "WARNING! Using --password via the CLI is insecure. Use --password-stdin.", "", -1)
			fmt.Println(out)
			return fmt.Errorf("Error authenticating base connector: exit status 1")
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
			return fmt.Errorf("error authenticating: exit status 1")
		}
	} else if p.Login.AccessToken != "" {
		cmd := commandLoginAccessToken(p.Login, p.Login.AccessToken)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error logging in to Docker registry: %s", err)
		}
		if strings.Contains(string(output), "Login Succeeded") {
			fmt.Println("Login successful")
		} else {
			return fmt.Errorf("login did not succeed")
		}
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

	// pre-pull cache images
	for _, img := range p.Build.CacheFrom {
		cmds = append(cmds, commandPull(img))
	}

	// setup for using ssh agent (https://docs.docker.com/develop/develop-images/build_enhancements/#using-ssh-to-access-private-data-in-builds)
	if p.Build.SSHAgentKey != "" {
		var sshErr error
		p.Build.SSHKeyPath, sshErr = writeSSHPrivateKey(p.Build.SSHAgentKey)
		if sshErr != nil {
			return sshErr
		}
	}

	cmds = append(cmds, commandBuild(p.Build)) // docker build

	for _, tag := range p.Build.Tags {
		cmds = append(cmds, commandTag(p.Build, tag)) // docker tag

		if !p.Dryrun {
			cmds = append(cmds, commandPush(p.Build, tag)) // docker push
		}
	}

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil && isCommandPull(cmd.Args) {
			fmt.Printf("Could not pull cache-from image %s. Ignoring...\n", cmd.Args[2])
		} else if err != nil && isCommandPrune(cmd.Args) {
			fmt.Printf("Could not prune system containers. Ignoring...\n")
		} else if err != nil && isCommandRmi(cmd.Args) {
			fmt.Printf("Could not remove image %s. Ignoring...\n", cmd.Args[2])
		} else if err != nil {
			return err
		}
	}

	// output the adaptive card
	if err := p.writeCard(); err != nil {
		fmt.Printf("Could not create adaptive card. %s\n", err)
	}

	if p.ArtifactFile != "" {
		if digest, err := getDigest(p.Build.TempTag); err == nil {
			if err = drone.WritePluginArtifactFile(p.Daemon.RegistryType, p.ArtifactFile, p.Daemon.Registry, p.Build.Repo, digest, p.Build.Tags); err != nil {
				fmt.Printf("failed to write plugin artifact file at path: %s with error: %s\n", p.ArtifactFile, err)
			}
		} else {
			fmt.Printf("Could not fetch the digest. %s\n", err)
		}
	}

	// execute cleanup routines in batch mode
	if p.Cleanup {
		// clear the slice
		cmds = nil

		cmds = append(cmds, commandRmi(p.Build.TempTag)) // docker rmi
		cmds = append(cmds, commandPrune())              // docker system prune -f

		for _, cmd := range cmds {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			trace(cmd)
		}
	}

	return nil
}

// helper function to set the credentials
func setDockerAuth(username, password, registry, baseImageUsername,
	baseImagePassword, baseImageRegistry string) ([]byte, error) {
	var credentials []docker.RegistryCredentials
	// add only docker registry to the config
	dockerConfig := docker.NewConfig()
	if password != "" {
		pushToRegistryCreds := docker.RegistryCredentials{
			Registry: registry,
			Username: username,
			Password: password,
		}
		// push registry auth
		credentials = append(credentials, pushToRegistryCreds)
	}

	if baseImageRegistry != "" {
		pullFromRegistryCreds := docker.RegistryCredentials{
			Registry: baseImageRegistry,
			Username: baseImageUsername,
			Password: baseImagePassword,
		}
		// base image registry auth
		credentials = append(credentials, pullFromRegistryCreds)
	}
	// Creates docker config for both the registries used for authentication
	return dockerConfig.CreateDockerConfigJson(credentials)
}

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

func commandLoginAccessToken(login Login, accessToken string) *exec.Cmd {
	cmd := exec.Command(dockerExe,
		"login",
		"-u",
		"oauth2accesstoken",
		"--password-stdin",
		login.Registry)
	cmd.Stdin = strings.NewReader(accessToken)
	return cmd
}

// helper to check if args match "docker pull <image>"
func isCommandPull(args []string) bool {
	return len(args) > 2 && args[1] == "pull"
}

func commandPull(repo string) *exec.Cmd {
	return exec.Command(dockerExe, "pull", repo)
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
		"-t", build.TempTag,
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
	if build.NoCache {
		args = append(args, "--no-cache")
	}
	for _, arg := range build.CacheFrom {
		args = append(args, "--cache-from", arg)
	}
	for _, arg := range build.ArgsEnv {
		addProxyValue(&build, arg)
	}
	for _, arg := range build.Args {
		args = append(args, "--build-arg", arg)
	}
	for _, host := range build.AddHost {
		args = append(args, "--add-host", host)
	}
	if build.Secret != "" {
		args = append(args, "--secret", build.Secret)
	}
	for _, secret := range build.SecretEnvs {
		if arg, err := getSecretStringCmdArg(secret); err == nil {
			args = append(args, "--secret", arg)
		}
	}
	for _, secret := range build.SecretFiles {
		if arg, err := getSecretFileCmdArg(secret); err == nil {
			args = append(args, "--secret", arg)
		}
	}
	if build.Target != "" {
		args = append(args, "--target", build.Target)
	}
	if build.Quiet {
		args = append(args, "--quiet")
	}
	if build.Platform != "" {
		args = append(args, "--platform", build.Platform)
	}
	if build.SSHKeyPath != "" {
		args = append(args, "--ssh", build.SSHKeyPath)
	}

	if build.AutoLabel {
		labelSchema := []string{
			fmt.Sprintf("created=%s", time.Now().Format(time.RFC3339)),
			fmt.Sprintf("revision=%s", build.Name),
			fmt.Sprintf("source=%s", build.Remote),
			fmt.Sprintf("url=%s", build.Link),
		}
		labelPrefix := "org.opencontainers.image"

		if len(build.LabelSchema) > 0 {
			labelSchema = append(labelSchema, build.LabelSchema...)
		}

		for _, label := range labelSchema {
			args = append(args, "--label", fmt.Sprintf("%s.%s", labelPrefix, label))
		}
	}

	if len(build.Labels) > 0 {
		for _, label := range build.Labels {
			args = append(args, "--label", label)
		}
	}

	// we need to enable buildkit, for secret support and ssh agent support
	if build.Secret != "" || len(build.SecretEnvs) > 0 || len(build.SecretFiles) > 0 || build.SSHAgentKey != "" {
		os.Setenv("DOCKER_BUILDKIT", "1")
	}
	return exec.Command(dockerExe, args...)
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

// helper function to create the docker tag command.
func commandTag(build Build, tag string) *exec.Cmd {
	var (
		source = build.TempTag
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

// helper to check if args match "docker prune"
func isCommandPrune(args []string) bool {
	return len(args) > 3 && args[2] == "prune"
}

func commandPrune() *exec.Cmd {
	return exec.Command(dockerExe, "system", "prune", "-f")
}

// helper to check if args match "docker rmi"
func isCommandRmi(args []string) bool {
	return len(args) > 2 && args[1] == "rmi"
}

func commandRmi(tag string) *exec.Cmd {
	return exec.Command(dockerExe, "rmi", tag)
}

func writeSSHPrivateKey(key string) (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %s", err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".ssh"), 0700); err != nil {
		return "", fmt.Errorf("unable to create .ssh directory: %s", err)
	}
	pathToKey := filepath.Join(home, ".ssh", "id_rsa")
	if err := os.WriteFile(pathToKey, []byte(key), 0400); err != nil {
		return "", fmt.Errorf("unable to write ssh key %s: %s", pathToKey, err)
	}
	path = fmt.Sprintf("default=%s", pathToKey)

	return path, nil
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}

func GetDroneDockerExecCmd() string {
	if runtime.GOOS == "windows" {
		return "C:/bin/drone-docker.exe"
	}

	return "drone-docker"
}

func getDigest(buildName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "--format='{{index .RepoDigests 0}}'", buildName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse the output to extract the repo digest.
	digest := strings.Trim(string(output), "'\n")
	parts := strings.Split(digest, "@")
	if len(parts) > 1 {
		return parts[1], nil
	}
	return "", errors.New("unable to fetch digest")
}
