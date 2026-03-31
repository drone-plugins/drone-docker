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
		RetryCount    int                // Number of retry attempts to reach Docker daemon
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
		Remote              string   // Git remote URL
		Name                string   // Docker build using default named tag
		TempTag             string   // Temporary tag used during docker build
		Dockerfile          string   // Docker build Dockerfile
		Context             string   // Docker build context
		Tags                []string // Docker build tags
		Args                []string // Docker build args
		ArgsEnv             []string // Docker build args from env
		ArgsNew             []string // docker build args which has comma seperated values
		IsMultipleBuildArgs bool     // env variable for fall back to old build args
		Target              string   // Docker build target
		Squash              bool     // Docker build squash
		Pull                bool     // Docker build pull
		CacheFrom           []string // Docker build cache-from
		Compress            bool     // Docker build compress
		Repo                string   // Docker build repository
		LabelSchema         []string // label-schema Label map
		AutoLabel           bool     // auto-label bool
		Labels              []string // Label map
		Link                string   // Git repo link
		NoCache             bool     // Docker build no-cache
		Secret              string   // secret keypair
		SecretEnvs          []string // Docker build secrets with env var as source
		SecretFiles         []string // Docker build secrets with file as source
		AddHost             []string // Docker build add-host
		Quiet               bool     // Docker build quiet
		Platform            string   // Docker build platform
		SSHAgentKey         string   // Docker build ssh agent key
		SSHKeyPath          string   // Docker build ssh key path
	}

	// CosignConfig defines Cosign signing parameters.
	CosignConfig struct {
		PrivateKey string // Private key content (PEM format) or file path
		Password   string // Password for encrypted private keys
		Params     string // Additional cosign parameters
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login             Login        // Docker login configuration
		Build             Build        // Docker build configuration
		Daemon            Daemon       // Docker daemon configuration
		Cosign            CosignConfig // Cosign signing configuration
		Dryrun            bool         // Docker push is skipped
		Cleanup           bool         // Docker purge is enabled
		CardPath          string       // Card path to write file to
		ArtifactFile      string       // Artifact path to write file to
		BaseImageRegistry string       // Docker registry to pull base image
		BaseImageUsername string       // Docker registry username to pull base image
		BaseImagePassword string       // Docker registry password to pull base image
		PushOnly          bool         // Push only mode, skips build process
		SourceImage       string       // Source image to push (optional)
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
	maxRetries := p.Daemon.RetryCount
	if maxRetries <= 0 {
		maxRetries = 15 // default value
	}
	for i := 0; ; i++ {
		cmd := commandInfo()
		err := cmd.Run()
		if err == nil {
			break
		}
		if i == maxRetries {
			fmt.Printf("Unable to reach Docker Daemon after %d attempts.\n", maxRetries)
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
			fmt.Printf("Username cannot be empty. The base image connector requires authenticated access. Please either use an authenticated connector, or remove the base image connector.")
		}
		if p.BaseImagePassword == "" {
			fmt.Printf("Password cannot be empty. The base image connector requires authenticated access. Please either use an authenticated connector, or remove the base image connector.")
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
	} else if !p.PushOnly {
		// Skip base image connector warning in push-only mode (not pulling anything)
		fmt.Println("\033[33mTo ensure consistent and reliable pipeline execution, we recommend setting up a Base Image Connector.\033[0m\n" +
			"\033[33mWhile optional at this time, configuring it helps prevent failures caused by Docker Hub's rate limits.\033[0m")
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

	// Enforce mutual exclusivity: push-only and dry-run cannot be used together
	if p.PushOnly && p.Dryrun {
		return fmt.Errorf("conflict: push-only and dry-run cannot be used together")
	}

	// Handle push-only mode if requested
	if p.PushOnly {
		return p.pushOnly()
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

	// Validate cosign configuration if present
	if p.shouldSignWithCosign() {
		if err := validateCosignConfig(p.Cosign); err != nil {
			return fmt.Errorf("cosign validation failed: %w", err)
		}
		fmt.Println("🔐 Cosign signing enabled - images will be signed after push")
	}

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

	// Handle cosign signing after all commands complete (like artifact generation)
	if p.shouldSignWithCosign() && !p.Dryrun {
		// Set up environment variables for cosign
		os.Setenv("COSIGN_YES", "true")

		if digest, err := getDigest(p.Build.TempTag); err == nil {
			fmt.Printf("🔐 Found image digest: %s\n", digest)
			
			// Sign with digest reference
			imageRef := fmt.Sprintf("%s@%s", p.Build.Repo, digest)
			cosignCmd := createCosignCommand(imageRef, p.Cosign)
			executeCosignCommand(cosignCmd)
		} else {
			fmt.Printf("⚠️  WARNING: Could not get image digest for cosign signing: %s\n", err)
			fmt.Printf("   Falling back to tag-based signing\n")
			
			// Fall back to tag-based signing for each tag
			for _, tag := range p.Build.Tags {
				imageRef := fmt.Sprintf("%s:%s", p.Build.Repo, tag)
				cosignCmd := createCosignCommand(imageRef, p.Cosign)
				executeCosignCommand(cosignCmd)
			}
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
	if build.IsMultipleBuildArgs {
		for _, arg := range build.ArgsNew {
			args = append(args, "--build-arg", arg)
		}
	} else {
		for _, arg := range build.Args {
			args = append(args, "--build-arg", arg)
		}
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
	if len(value) > 0 && !hasProxyBuildArgNew(build, key) {
		build.ArgsNew = append(build.ArgsNew, fmt.Sprintf("%s=%s", key, value))
		build.ArgsNew = append(build.ArgsNew, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}
}

// helper function to get a proxy value from the environment.
//
// Checks in order: lowercase key, uppercase key, then HARNESS_<UPPERCASE_KEY>.
// Assumes that the upper and lower case versions are the same value.
func getProxyValue(key string) string {
	value := os.Getenv(key)

	if len(value) > 0 {
		return value
	}

	value = os.Getenv(strings.ToUpper(key))

	if len(value) > 0 {
		return value
	}

	harnessValue := os.Getenv("HARNESS_" + strings.ToUpper(key))
	if len(harnessValue) > 0 {
		fmt.Printf("Using HARNESS_%s as proxy value for %s\n", strings.ToUpper(key), key)
	}
	return harnessValue
}

// helper function that looks to see if a proxy value was set in the build args.
func hasProxyBuildArg(build *Build, key string) bool {
	keyUpper := strings.ToUpper(key)
	harnessKey := "HARNESS_" + keyUpper

	for _, s := range build.Args {
		if strings.HasPrefix(s, key) || strings.HasPrefix(s, keyUpper) || strings.HasPrefix(s, harnessKey) {
			return true
		}
	}

	return false
}
func hasProxyBuildArgNew(build *Build, key string) bool {
	keyUpper := strings.ToUpper(key)
	harnessKey := "HARNESS_" + keyUpper

	for _, s := range build.ArgsNew {
		if strings.HasPrefix(s, key) || strings.HasPrefix(s, keyUpper) || strings.HasPrefix(s, harnessKey) {
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

// helper to check if args match "cosign sign"
func isCommandCosign(args []string) bool {
	return len(args) > 1 && args[0] == cosignExe
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
	cmd := exec.Command(dockerExe, "inspect", "--format='{{index .RepoDigests 0}}'", buildName)
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

// imageExists checks if an image exists in local daemon
func imageExists(tag string) bool {
	cmd := exec.Command(dockerExe, "image", "inspect", tag)
	return cmd.Run() == nil
}

// getDigestAfterPush gets digest from a pushed image
func getDigestAfterPush(tag string) (string, error) {
	cmd := exec.Command(dockerExe, "inspect", "--format", "{{ index (split (index .RepoDigests 0) \"@\") 1 }}", tag)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get digest for %s: %w", tag, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// shouldSignWithCosign determines if cosign signing should be performed
func (p Plugin) shouldSignWithCosign() bool {
	return p.Cosign.PrivateKey != ""
}

// validateCosignConfig validates the cosign configuration
func validateCosignConfig(config CosignConfig) error {
	if config.PrivateKey == "" {
		return nil // No cosign config, skip silently
	}

	// Check if cosign binary is available
	if _, err := exec.LookPath(cosignExe); err != nil {
		fmt.Printf("❌ ERROR: cosign binary not found at %s\n", cosignExe)
		fmt.Println("   Ensure you're using a plugin image that includes cosign")
		return fmt.Errorf("cosign binary not available: %w", err)
	}

	// Check if it's trying to use keyless signing
	if strings.Contains(config.Params, "--oidc") ||
		strings.Contains(config.Params, "--identity-token") {
		fmt.Println("⚠️  WARNING: Keyless signing (OIDC) isn't supported yet in this plugin. Use private key signing instead.")
		return errors.New("keyless signing not supported")
	}

	// Validate private key format if it's PEM content
	if strings.HasPrefix(config.PrivateKey, "-----BEGIN") {
		if !isValidPEMKey(config.PrivateKey) {
			return errors.New("❌ Invalid private key format. Expected PEM format")
		}

		// Check encrypted key password requirement
		if isEncryptedPEMKey(config.PrivateKey) && config.Password == "" {
			return errors.New("🔐 Encrypted private key requires password. Set PLUGIN_COSIGN_PASSWORD")
		}

	} else {
		// File-based key - check if it's accessible (basic check)
		if _, err := os.Stat(config.PrivateKey); err != nil {
			fmt.Printf("⚠️  WARNING: Private key file may not be accessible: %s\n", config.PrivateKey)
			fmt.Println("   This will be verified during signing")
		}
	}

	return nil
}

// isEncryptedPEMKey checks if a PEM key is encrypted
func isEncryptedPEMKey(pemContent string) bool {
	return strings.Contains(pemContent, "ENCRYPTED")
}

// isValidPEMKey performs basic PEM format validation
func isValidPEMKey(pemContent string) bool {
	return strings.Contains(pemContent, "-----BEGIN") &&
		strings.Contains(pemContent, "-----END") &&
		(strings.Contains(pemContent, "PRIVATE KEY") ||
			strings.Contains(pemContent, "RSA PRIVATE KEY") ||
			strings.Contains(pemContent, "EC PRIVATE KEY"))
}

// createCosignCommand creates a cosign sign command with the given image reference
func createCosignCommand(imageRef string, cosign CosignConfig) *exec.Cmd {
	args := []string{"sign", "--yes"}
	
	// Handle private key (content vs file path)
	if strings.HasPrefix(cosign.PrivateKey, "-----BEGIN") {
		args = append(args, "--key", "env://COSIGN_PRIVATE_KEY")
		os.Setenv("COSIGN_PRIVATE_KEY", cosign.PrivateKey)
	} else {
		args = append(args, "--key", cosign.PrivateKey)
	}
	
	// Set password if provided
	if cosign.Password != "" {
		os.Setenv("COSIGN_PASSWORD", cosign.Password)
	}
	
	// Add any extra parameters
	if cosign.Params != "" {
		extraArgs := strings.Fields(cosign.Params)
		args = append(args, extraArgs...)
	}
	
	// Add the image reference to sign
	args = append(args, imageRef)
	
	return exec.Command(cosignExe, args...)
}

// executeCosignCommand executes the given cosign command and handles errors
func executeCosignCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("🚀 Executing: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))

	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  WARNING: Image signing failed: %s\n", err)
		fmt.Printf("   Image was pushed successfully but could not be signed\n")
		fmt.Printf("   This is not fatal - continuing with the build\n")
	}
}

// pushOnly handles pushing images without building them
func (p Plugin) pushOnly() error {
	// Check if source image is specified
	sourceImageName := p.SourceImage
	var sourceTags []string

	if sourceImageName == "" {
		// If no source image specified, use the repo and first tag
		fmt.Println("source_image not provided, using repo and tag value")
		sourceImageName = p.Build.Repo
		sourceTags = p.Build.Tags
	} else {
		// If source image is specified, check if it has a tag
		lastColonIndex := strings.LastIndex(sourceImageName, ":")
		if lastColonIndex > 0 && lastColonIndex < len(sourceImageName) {
			// Check if there's a slash after the last colon (indicating it's a port, not a tag)
			// For example: registry:5000/image (has slash after colon - port not tag)
			// vs image:tag (no slash after colon - it's a tag)
			if strings.LastIndex(sourceImageName, "/") > lastColonIndex {
				// The last colon is part of the registry:port, not a tag separator
				sourceTags = []string{"latest"}
			} else {
				// The last colon separates the tag
				tag := sourceImageName[lastColonIndex+1:]
				sourceImageName = sourceImageName[:lastColonIndex]

				if tag == "" {
					fmt.Printf("No tag specified in source image (or empty tag). Using 'latest' as the default tag.\n")
					tag = "latest"
				}
				sourceTags = []string{tag}
			}
		} else {
			// Default to "latest" if no tag specified
			sourceTags = []string{"latest"}
		}
		fmt.Printf("Using source image: %s with tag(s): %s\n", sourceImageName, strings.Join(sourceTags, ", "))
	}

	// For each source tag and target tag combination
	var digest string
	var firstPushedImage string

	for _, sourceTag := range sourceTags {
		sourceFullImageName := fmt.Sprintf("%s:%s", sourceImageName, sourceTag)

		// Check if the source image exists in local daemon
		if !imageExists(sourceFullImageName) {
			fmt.Printf("Warning: Source image %s not found\n", sourceFullImageName)
			// Continue to the next source tag if available, otherwise return error
			if len(sourceTags) > 1 {
				continue
			}
			return fmt.Errorf("source image %s not found, cannot push", sourceFullImageName)
		}

		// For each target tag, tag and push
		for _, targetTag := range p.Build.Tags {
			targetFullImageName := fmt.Sprintf("%s:%s", p.Build.Repo, targetTag)

			// Skip if source and target are identical
			if sourceFullImageName == targetFullImageName {
				fmt.Printf("Source and target image names are identical: %s\n", sourceFullImageName)
			} else {
				// Tag the source image with the target name
				fmt.Printf("Tagging %s as %s\n", sourceFullImageName, targetFullImageName)
				tagCmd := exec.Command(dockerExe, "tag", sourceFullImageName, targetFullImageName)
				tagCmd.Stdout = os.Stdout
				tagCmd.Stderr = os.Stderr
				trace(tagCmd)
				if err := tagCmd.Run(); err != nil {
					return fmt.Errorf("failed to tag image %s as %s: %w", sourceFullImageName, targetFullImageName, err)
				}
			}
		}
	}

	// Push all target images
	for _, tag := range p.Build.Tags {
		fullImageName := fmt.Sprintf("%s:%s", p.Build.Repo, tag)

		// Check if image exists in local daemon
		if !imageExists(fullImageName) {
			return fmt.Errorf("image %s not found, cannot push", fullImageName)
		}

		// Push image
		fmt.Println("Pushing image:", fullImageName)
		pushCmd := commandPush(p.Build, tag)
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		trace(pushCmd)
		if err := pushCmd.Run(); err != nil {
			return fmt.Errorf("failed to push image %s: %w", fullImageName, err)
		}

		// Track the first pushed image for card generation
		if firstPushedImage == "" {
			firstPushedImage = fullImageName
		}

		// Get the digest after push (we only need one)
		if digest == "" {
			d, err := getDigestAfterPush(fullImageName)
			if err == nil {
				digest = d
			} else {
				fmt.Printf("Warning: Could not get digest for %s: %v\n", fullImageName, err)
			}
		}
	}

	// Output the adaptive card
	if firstPushedImage != "" {
		if err := p.writeCardForImage(firstPushedImage); err != nil {
			fmt.Printf("Could not create adaptive card. %s\n", err)
		}
	}

	// Write to artifact file
	if p.ArtifactFile != "" && digest != "" {
		if err := drone.WritePluginArtifactFile(
			p.Daemon.RegistryType,
			p.ArtifactFile,
			p.Daemon.Registry,
			p.Build.Repo,
			digest,
			p.Build.Tags,
		); err != nil {
			fmt.Printf("Failed to write plugin artifact file at path: %s with error: %s\n",
				p.ArtifactFile, err)
		}
	}

	// Handle cosign signing after push
	if p.shouldSignWithCosign() {
		// Set up environment variables for cosign
		os.Setenv("COSIGN_YES", "true")

		if digest != "" {
			fmt.Printf("🔐 Found image digest: %s\n", digest)

			// Sign with digest reference
			imageRef := fmt.Sprintf("%s@%s", p.Build.Repo, digest)
			cosignCmd := createCosignCommand(imageRef, p.Cosign)
			executeCosignCommand(cosignCmd)
		} else {
			fmt.Printf("⚠️  WARNING: Could not get image digest for cosign signing\n")
			fmt.Printf("   Falling back to tag-based signing\n")

			// Fall back to tag-based signing for each tag
			for _, tag := range p.Build.Tags {
				imageRef := fmt.Sprintf("%s:%s", p.Build.Repo, tag)
				cosignCmd := createCosignCommand(imageRef, p.Cosign)
				executeCosignCommand(cosignCmd)
			}
		}
	}

	return nil
}
