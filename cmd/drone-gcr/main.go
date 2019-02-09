package main

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path"
	"strings"
)

// gcr default username
const username = "_json_key"

func main() {
	var (
		repo     = getenv("PLUGIN_REPO")
		registry = getenv("PLUGIN_REGISTRY")
		password = getenv(
			"PLUGIN_JSON_KEY",
			"GCR_JSON_KEY",
			"GOOGLE_CREDENTIALS",
			"TOKEN",
		)
	)

	// decode the token if base64 encoded
	decoded, err := base64.StdEncoding.DecodeString(password)
	if err == nil {
		password = string(decoded)
	}

	// default registry value
	if registry == "" {
		registry = "gcr.io"
	}

	// must use the fully qualified repo name. If the
	// repo name does not have the registry prefix we
	// should prepend.
	if !strings.HasPrefix(repo, registry) {
		repo = path.Join(registry, repo)
	}

	os.Setenv("PLUGIN_REPO", repo)
	os.Setenv("PLUGIN_REGISTRY", registry)
	os.Setenv("DOCKER_USERNAME", username)
	os.Setenv("DOCKER_PASSWORD", password)

	// invoke the base docker plugin binary
	cmd := exec.Command("drone-docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
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
