package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var build = "0" // build number set at compile-time

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "docker plugin"
	app.Usage = "docker plugin"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run disables docker push",
			EnvVar: "PLUGIN_DRY_RUN",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
			Value:  "00000000",
		},
		cli.StringFlag{
			Name:   "daemon.mirror",
			Usage:  "docker daemon registry mirror",
			EnvVar: "PLUGIN_MIRROR",
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
			EnvVar: "PLUGIN_DNS",
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
			Usage:  "docker daemon executes in debug mode",
			EnvVar: "PLUGIN_DAEMON_OFF",
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
			Name:   "tags",
			Usage:  "build tags",
			Value:  &cli.StringSlice{"latest"},
			EnvVar: "PLUGIN_TAG,PLUGIN_TAGS",
		},
		cli.StringSliceFlag{
			Name:   "args",
			Usage:  "build args",
			EnvVar: "PLUGIN_BUILD_ARGS",
		},
		cli.BoolFlag{
			Name:   "squash",
			Usage:  "squash the layers at build time",
			EnvVar: "PLUGIN_SQUASH",
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
		cli.StringFlag{
			Name:   "docker.registry",
			Usage:  "docker registry",
			Value:  defaultRegistry,
			EnvVar: "DOCKER_REGISTRY,PLUGIN_REGISTRY",
		},
		cli.StringFlag{
			Name:   "docker.username",
			Usage:  "docker username",
			EnvVar: "DOCKER_USERNAME,PLUGIN_USERNAME",
		},
		cli.StringFlag{
			Name:   "docker.password",
			Usage:  "docker password",
			EnvVar: "DOCKER_PASSWORD,PLUGIN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "docker.email",
			Usage:  "docker email",
			EnvVar: "DOCKER_EMAIL,PLUGIN_EMAIL",
		},
		cli.StringFlag{
			Name:   "push.registries",
			Usage:  "push registries",
			EnvVar: "PLUGIN_PUSH_REGISTRIES",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := Plugin{
		Dryrun: c.Bool("dry-run"),
		Login: Login{
			Registry: c.String("docker.registry"),
			Username: c.String("docker.username"),
			Password: c.String("docker.password"),
			Email:    c.String("docker.email"),
		},
		Build: Build{
			Name:       c.String("commit.sha"),
			Dockerfile: c.String("dockerfile"),
			Context:    c.String("context"),
			Tags:       c.StringSlice("tags"),
			Args:       c.StringSlice("args"),
			Squash:     c.Bool("squash"),
			Compress:   c.Bool("compress"),
			Repo:       c.String("repo"),
		},
		Daemon: Daemon{
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
			MTU:           c.String("daemon.mtu"),
			Experimental:  c.Bool("daemon.experimental"),
		},
	}

	if r, err := parsePushRegistries(c.String("push.registries"), plugin.Login.Registry); err == nil {
		plugin.PushRegistries = r
	} else {
		return err
	}

	return plugin.Exec()
}

func parsePushRegistries(raw, defaultRegistry string) ([]Login, error) {
	if raw == "" {
		return nil, nil
	}

	var out []Login

	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("Failed to parse push registries: %s", err.Error())
	}

	// Default to the main registry set, in case only a repo is provided
	for i := range out {
		if out[i].Registry == "" {
			out[i].Registry = defaultRegistry
		}
	}

	return out, nil
}
