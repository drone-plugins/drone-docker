package main

import (
	"os"
	"os/exec"
	"path"
)

func main() {
	var (
		registry = "registry.heroku.com"
		process  = os.Getenv("PLUGIN_PROCESS_TYPE")
		app      = os.Getenv("PLUGIN_APP")
		email    = os.Getenv("PLUGIN_EMAIL")
		key      = os.Getenv("PLUGIN_API_KEY")
	)

	// if the heroku email is provided as a named secret
	// then we should use it.
	if os.Getenv("HEROKU_EMAIL") != "" {
		email = os.Getenv("HEROKU_EMAIL")
	}

	// if the heroku api key is provided as a named secret
	// then we should use it.
	if os.Getenv("HEROKU_API_KEY") != "" {
		key = os.Getenv("HEROKU_API_KEY")
	}

	os.Setenv("PLUGIN_REGISTRY", registry)
	os.Setenv("PLUGIN_REPO", path.Join(registry, app, process))

	os.Setenv("DOCKER_PASSWORD", key)
	os.Setenv("DOCKER_USERNAME", email)
	os.Setenv("DOCKER_EMAIL", email)

	// invoke the base docker plugin binary
	cmd := exec.Command("drone-docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}
