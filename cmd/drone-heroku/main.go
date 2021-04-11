package main

import (
	"os"
	"os/exec"
	"path"

	"github.com/joho/godotenv"
)

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	var (
		registry  = "registry.heroku.com"
		process   = getenv("PLUGIN_PROCESS_TYPE")
		app       = getenv("PLUGIN_APP")
		email     = getenv("PLUGIN_EMAIL", "HEROKU_EMAIL")
		key       = getenv("PLUGIN_API_KEY", "HEROKU_API_KEY")
		autolabel = getenv("PLUGIN_AUTO_LABEL")
	)

	if process == "" {
		process = "web"
	}

	os.Setenv("PLUGIN_REGISTRY", registry)
	os.Setenv("PLUGIN_REPO", path.Join(registry, app, process))
	os.Setenv("PLUGIN_AUTO_LABEL", autolabel)

	os.Setenv("DOCKER_PASSWORD", key)
	os.Setenv("DOCKER_USERNAME", email)
	os.Setenv("DOCKER_EMAIL", email)

	cmd := exec.Command("drone-docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

func getenv(key ...string) (s string) {
	for _, k := range key {
		s = os.Getenv(k)
		if s != "" {
			return
		}
	}
	return
}
