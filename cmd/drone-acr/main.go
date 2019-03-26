package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var (
		repo     = getenv("PLUGIN_REPO")
		registry = getenv("PLUGIN_REGISTRY")
		username = getenv("SERVICE_PRINCIPAL_CLIENT_ID")
		password = getenv("SERVICE_PRINCIPAL_CLIENT_SECRET")
	)

	// default registry value
	if registry == "" {
		registry = "azurecr.io"
	}

	// must use the fully qualified repo name. If the
	// repo name does not have the registry prefix we
	// should prepend.
	if !strings.HasPrefix(repo, registry) {
		repo = fmt.Sprintf("%s/%s", registry, repo)
	}

	os.Setenv("PLUGIN_REPO", repo)
	os.Setenv("PLUGIN_REGISTRY", registry)
	os.Setenv("DOCKER_USERNAME", username)
	os.Setenv("DOCKER_PASSWORD", password)

	// invoke the base docker plugin binary
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
