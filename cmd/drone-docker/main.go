package main

import (
	"os"
	"runtime"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	docker "github.com/drone-plugins/drone-docker"
	"github.com/drone-plugins/drone-plugin-lib/drone"
)

var (
	version = "unknown"
)

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "docker plugin"
	app.Usage = "docker plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run disables docker push",
			EnvVar: "PLUGIN_DRY_RUN, PLUGIN_NO_PUSH",
		},
		cli.StringFlag{
			Name:   "remote.url",
			Usage:  "git remote url",
			EnvVar: "DRONE_REMOTE_URL",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
			Value:  "00000000",
		},
		cli.StringFlag{
			Name:   "commit.ref",
			Usage:  "git commit ref",
			EnvVar: "DRONE_COMMIT_REF",
		},
		cli.StringFlag{
			Name:   "daemon.mirror",
			Usage:  "docker daemon registry mirror",
			EnvVar: "PLUGIN_MIRROR,DOCKER_PLUGIN_MIRROR",
		},
		cli.StringFlag{
			Name:   "daemon.storage-driver",
			Usage:  "docker daemon storage driver",
			EnvVar: "PLUGIN_STORAGE_DRIVER",
		},
		cli.StringFlag{
			Name:   "daemon.storage-path",
			Usage:  "docker daemon storage path",
			Value:  "/var/lib/docker",
			EnvVar: "PLUGIN_STORAGE_PATH",
		},
		cli.StringFlag{
			Name:   "daemon.bip",
			Usage:  "docker daemon bride ip address",
			EnvVar: "PLUGIN_BIP",
		},
		cli.StringFlag{
			Name:   "daemon.mtu",
			Usage:  "docker daemon custom mtu setting",
			EnvVar: "PLUGIN_MTU",
		},
		cli.StringSliceFlag{
			Name:   "daemon.dns",
			Usage:  "docker daemon dns server",
			EnvVar: "PLUGIN_CUSTOM_DNS",
		},
		cli.StringSliceFlag{
			Name:   "daemon.dns-search",
			Usage:  "docker daemon dns search domains",
			EnvVar: "PLUGIN_CUSTOM_DNS_SEARCH",
		},
		cli.BoolFlag{
			Name:   "daemon.insecure",
			Usage:  "docker daemon allows insecure registries",
			EnvVar: "PLUGIN_INSECURE",
		},
		cli.BoolFlag{
			Name:   "daemon.ipv6",
			Usage:  "docker daemon IPv6 networking",
			EnvVar: "PLUGIN_IPV6",
		},
		cli.BoolFlag{
			Name:   "daemon.experimental",
			Usage:  "docker daemon Experimental mode",
			EnvVar: "PLUGIN_EXPERIMENTAL",
		},
		cli.BoolFlag{
			Name:   "daemon.debug",
			Usage:  "docker daemon executes in debug mode",
			EnvVar: "PLUGIN_DEBUG,DOCKER_LAUNCH_DEBUG",
		},
		cli.BoolFlag{
			Name:   "daemon.off",
			Usage:  "don't start the docker daemon",
			EnvVar: "PLUGIN_DAEMON_OFF",
		},
		cli.IntFlag{
			Name:   "daemon.retry-count",
			Usage:  "number of retry attempts to reach docker daemon",
			Value:  15,
			EnvVar: "PLUGIN_DAEMON_RETRY_COUNT",
		},
		cli.StringFlag{
			Name:   "dockerfile",
			Usage:  "build dockerfile",
			Value:  "Dockerfile",
			EnvVar: "PLUGIN_DOCKERFILE",
		},
		cli.StringFlag{
			Name:   "context",
			Usage:  "build context",
			Value:  ".",
			EnvVar: "PLUGIN_CONTEXT",
		},
		cli.StringSliceFlag{
			Name:     "tags",
			Usage:    "build tags",
			Value:    &cli.StringSlice{"latest"},
			EnvVar:   "PLUGIN_TAG,PLUGIN_TAGS",
			FilePath: ".tags",
		},
		cli.BoolFlag{
			Name:   "tags.auto",
			Usage:  "default build tags",
			EnvVar: "PLUGIN_DEFAULT_TAGS,PLUGIN_AUTO_TAG",
		},
		cli.StringFlag{
			Name:   "tags.suffix",
			Usage:  "default build tags with suffix",
			EnvVar: "PLUGIN_DEFAULT_SUFFIX,PLUGIN_AUTO_TAG_SUFFIX",
		},
		cli.StringSliceFlag{
			Name:   "args",
			Usage:  "build args",
			EnvVar: "PLUGIN_BUILD_ARGS",
		},
		cli.StringSliceFlag{
			Name:   "args-from-env",
			Usage:  "build args",
			EnvVar: "PLUGIN_BUILD_ARGS_FROM_ENV",
		},
		cli.GenericFlag{
			Name:   "args-new",
			Usage:  "build args new",
			EnvVar: "PLUGIN_BUILD_ARGS_NEW",
			Value:  new(CustomStringSliceFlag),
		},
		cli.BoolFlag{
			Name:   "plugin-multiple-build-agrs",
			Usage:  "plugin multiple build agrs",
			EnvVar: "PLUGIN_MULTIPLE_BUILD_ARGS",
		},
		cli.BoolFlag{
			Name:   "quiet",
			Usage:  "quiet docker build",
			EnvVar: "PLUGIN_QUIET",
		},
		cli.StringFlag{
			Name:   "target",
			Usage:  "build target",
			EnvVar: "PLUGIN_TARGET",
		},
		cli.StringSliceFlag{
			Name:   "cache-from",
			Usage:  "images to consider as cache sources",
			EnvVar: "PLUGIN_CACHE_FROM",
		},
		cli.BoolFlag{
			Name:   "squash",
			Usage:  "squash the layers at build time",
			EnvVar: "PLUGIN_SQUASH",
		},
		cli.BoolTFlag{
			Name:   "pull-image",
			Usage:  "force pull base image at build time",
			EnvVar: "PLUGIN_PULL_IMAGE",
		},
		cli.BoolFlag{
			Name:   "compress",
			Usage:  "compress the build context using gzip",
			EnvVar: "PLUGIN_COMPRESS",
		},
		cli.StringFlag{
			Name:   "repo",
			Usage:  "docker repository",
			EnvVar: "PLUGIN_REPO",
		},
		cli.StringSliceFlag{
			Name:   "custom-labels",
			Usage:  "additional k=v labels",
			EnvVar: "PLUGIN_CUSTOM_LABELS",
		},
		cli.StringSliceFlag{
			Name:   "label-schema",
			Usage:  "label-schema labels",
			EnvVar: "PLUGIN_LABEL_SCHEMA",
		},
		cli.BoolTFlag{
			Name:   "auto-label",
			Usage:  "auto-label true|false",
			EnvVar: "PLUGIN_AUTO_LABEL",
		},
		cli.StringFlag{
			Name:   "link",
			Usage:  "link https://example.com/org/repo-name",
			EnvVar: "PLUGIN_REPO_LINK,DRONE_REPO_LINK",
		},
		cli.StringFlag{
			Name:   "docker.registry",
			Usage:  "docker registry",
			Value:  "https://index.docker.io/v1/",
			EnvVar: "PLUGIN_REGISTRY,DOCKER_REGISTRY",
		},
		cli.StringFlag{
			Name:   "docker.username",
			Usage:  "docker username",
			EnvVar: "PLUGIN_USERNAME,DOCKER_USERNAME",
		},
		cli.StringFlag{
			Name:   "docker.password",
			Usage:  "docker password",
			EnvVar: "PLUGIN_PASSWORD,DOCKER_PASSWORD",
		},
		cli.StringFlag{
			Name:   "docker.baseimageusername",
			Usage:  "Docker username for base image registry",
			EnvVar: "PLUGIN_DOCKER_USERNAME,PLUGIN_BASE_IMAGE_USERNAME,DOCKER_BASE_IMAGE_USERNAME",
		},
		cli.StringFlag{
			Name:   "docker.baseimagepassword",
			Usage:  "Docker password for base image registry",
			EnvVar: "PLUGIN_DOCKER_PASSWORD,PLUGIN_BASE_IMAGE_PASSWORD,DOCKER_BASE_IMAGE_PASSWORD",
		},
		cli.StringFlag{
			Name:   "docker.baseimageregistry",
			Usage:  "Docker registry for base image registry",
			EnvVar: "PLUGIN_DOCKER_REGISTRY,PLUGIN_BASE_IMAGE_REGISTRY,DOCKER_BASE_IMAGE_REGISTRY",
		},
		cli.StringFlag{
			Name:   "docker.email",
			Usage:  "docker email",
			EnvVar: "PLUGIN_EMAIL,DOCKER_EMAIL",
		},
		cli.StringFlag{
			Name:   "docker.config",
			Usage:  "docker json dockerconfig content",
			EnvVar: "PLUGIN_CONFIG,DOCKER_PLUGIN_CONFIG",
		},
		cli.BoolTFlag{
			Name:   "docker.purge",
			Usage:  "docker should cleanup images",
			EnvVar: "PLUGIN_PURGE",
		},
		cli.StringFlag{
			Name:   "repo.branch",
			Usage:  "repository default branch",
			EnvVar: "DRONE_REPO_BRANCH",
		},
		cli.BoolFlag{
			Name:   "no-cache",
			Usage:  "do not use cached intermediate containers",
			EnvVar: "PLUGIN_NO_CACHE",
		},
		cli.StringSliceFlag{
			Name:   "add-host",
			Usage:  "additional host:IP mapping",
			EnvVar: "PLUGIN_ADD_HOST",
		},
		cli.StringFlag{
			Name:   "secret",
			Usage:  "secret key value pair eg id=MYSECRET",
			EnvVar: "PLUGIN_SECRET",
		},
		cli.StringSliceFlag{
			Name:   "secrets-from-env",
			Usage:  "secret key value pair eg secret_name=secret",
			EnvVar: "PLUGIN_SECRETS_FROM_ENV",
		},
		cli.StringSliceFlag{
			Name:   "secrets-from-file",
			Usage:  "secret key value pairs eg secret_name=/path/to/secret",
			EnvVar: "PLUGIN_SECRETS_FROM_FILE",
		},
		cli.StringFlag{
			Name:   "drone-card-path",
			Usage:  "card path location to write to",
			EnvVar: "DRONE_CARD_PATH",
		},
		cli.StringFlag{
			Name:   "platform",
			Usage:  "platform value to pass to docker",
			EnvVar: "PLUGIN_PLATFORM",
		},
		cli.StringFlag{
			Name:   "ssh-agent-key",
			Usage:  "ssh agent key to use",
			EnvVar: "PLUGIN_SSH_AGENT_KEY",
		},
		cli.StringFlag{
			Name:   "artifact-file",
			Usage:  "Artifact file location that will be generated by the plugin. This file will include information of docker images that are uploaded by the plugin.",
			EnvVar: "PLUGIN_ARTIFACT_FILE",
		},
		cli.StringFlag{
			Name:   "registry-type",
			Usage:  "registry type",
			EnvVar: "PLUGIN_REGISTRY_TYPE",
		},
		cli.StringFlag{
			Name:   "access-token",
			Usage:  "access token",
			EnvVar: "ACCESS_TOKEN",
		},
		// Cosign signing configuration
		cli.StringFlag{
			Name:   "cosign.private-key",
			Usage:  "cosign private key content or file path for signing",
			EnvVar: "PLUGIN_COSIGN_PRIVATE_KEY",
		},
		cli.StringFlag{
			Name:   "cosign.password",
			Usage:  "password for encrypted cosign private key",
			EnvVar: "PLUGIN_COSIGN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "cosign.params",
			Usage:  "additional cosign parameters (e.g., annotations, flags)",
			EnvVar: "PLUGIN_COSIGN_PARAMS",
		},
		cli.BoolFlag{
			Name:   "push-only",
			Usage:  "skip build and only push images",
			EnvVar: "PLUGIN_PUSH_ONLY",
		},
		cli.StringFlag{
			Name:   "source-image",
			Usage:  "source image to tag and push (format: repo:tag)",
			EnvVar: "PLUGIN_SOURCE_IMAGE",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	registryType := drone.Docker
	if c.String("registry-type") != "" {
		registryType = drone.RegistryType(c.String("registry-type"))
	}

	plugin := docker.Plugin{
		Dryrun:  c.Bool("dry-run"),
		Cleanup: c.BoolT("docker.purge"),
		Login: docker.Login{
			Registry:    c.String("docker.registry"),
			Username:    c.String("docker.username"),
			Password:    c.String("docker.password"),
			Email:       c.String("docker.email"),
			Config:      c.String("docker.config"),
			AccessToken: c.String("access-token"),
		},
		CardPath:     c.String("drone-card-path"),
		ArtifactFile: c.String("artifact-file"),
		Build: docker.Build{
			Remote:              c.String("remote.url"),
			Name:                c.String("commit.sha"),
			TempTag:             generateTempTag(),
			Dockerfile:          c.String("dockerfile"),
			Context:             c.String("context"),
			Tags:                c.StringSlice("tags"),
			Args:                c.StringSlice("args"),
			ArgsEnv:             c.StringSlice("args-from-env"),
			ArgsNew:             c.Generic("args-new").(*CustomStringSliceFlag).GetValue(),
			IsMultipleBuildArgs: c.Bool("plugin-multiple-build-agrs"),
			Target:              c.String("target"),
			Squash:              c.Bool("squash"),
			Pull:                c.BoolT("pull-image"),
			CacheFrom:           c.StringSlice("cache-from"),
			Compress:            c.Bool("compress"),
			Repo:                c.String("repo"),
			Labels:              c.StringSlice("custom-labels"),
			LabelSchema:         c.StringSlice("label-schema"),
			AutoLabel:           c.BoolT("auto-label"),
			Link:                c.String("link"),
			NoCache:             c.Bool("no-cache"),
			Secret:              c.String("secret"),
			SecretEnvs:          c.StringSlice("secrets-from-env"),
			SecretFiles:         c.StringSlice("secrets-from-file"),
			AddHost:             c.StringSlice("add-host"),
			Quiet:               c.Bool("quiet"),
			Platform:            c.String("platform"),
			SSHAgentKey:         c.String("ssh-agent-key"),
		},
		Daemon: docker.Daemon{
			Registry:      c.String("docker.registry"),
			Mirror:        c.String("daemon.mirror"),
			StorageDriver: c.String("daemon.storage-driver"),
			StoragePath:   c.String("daemon.storage-path"),
			Insecure:      c.Bool("daemon.insecure"),
			Disabled:      c.Bool("daemon.off"),
			IPv6:          c.Bool("daemon.ipv6"),
			Debug:         c.Bool("daemon.debug"),
			Bip:           c.String("daemon.bip"),
			DNS:           c.StringSlice("daemon.dns"),
			DNSSearch:     c.StringSlice("daemon.dns-search"),
			MTU:           c.String("daemon.mtu"),
			Experimental:  c.Bool("daemon.experimental"),
			RetryCount:    c.Int("daemon.retry-count"),
			RegistryType:  registryType,
		},
		BaseImageRegistry: c.String("docker.baseimageregistry"),
		BaseImageUsername: c.String("docker.baseimageusername"),
		BaseImagePassword: c.String("docker.baseimagepassword"),
		Cosign: docker.CosignConfig{
			PrivateKey: c.String("cosign.private-key"),
			Password:   c.String("cosign.password"),
			Params:     c.String("cosign.params"),
		},
		PushOnly:    c.Bool("push-only"),
		SourceImage: c.String("source-image"),
	}

	if c.Bool("tags.auto") {
		if docker.UseDefaultTag( // return true if tag event or default branch
			c.String("commit.ref"),
			c.String("repo.branch"),
		) {
			tag, err := docker.DefaultTagSuffix(
				c.String("commit.ref"),
				c.String("tags.suffix"),
			)
			if err != nil {
				logrus.Printf("cannot build docker image for %s, invalid semantic version", c.String("commit.ref"))
				return err
			}
			plugin.Build.Tags = tag
		} else {
			logrus.Printf("skipping automated docker build for %s", c.String("commit.ref"))
			return nil
		}
	}

	return plugin.Exec()
}

func generateTempTag() string {
	return strings.ToLower(uniuri.New())
}

func GetExecCmd() string {
	if runtime.GOOS == "windows" {
		return "C:/bin/drone-docker.exe"
	}

	return "drone-docker"
}
