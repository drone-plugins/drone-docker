package main

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"

	docker "github.com/drone-plugins/drone-docker"
)

// gcr default username
var username = "_json_key"

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	var (
		repo     = getenv("PLUGIN_REPO")
		registry = getenv("PLUGIN_REGISTRY")
		password = getenv(
			"PLUGIN_JSON_KEY",
			"GCR_JSON_KEY",
			"GOOGLE_CREDENTIALS",
			"TOKEN",
		)
		is_workload_identity = parseBoolOrDefault(false, getenv("PLUGIN_IS_WORKLOAD_IDENTITY"))
	)
	// set username and password
	username, password = setUsernameAndPassword(username, password, is_workload_identity)
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
	os.Setenv("PLUGIN_REGISTRY_TYPE", "GCR")

	// invoke the base docker plugin binary
	cmd := exec.Command(docker.GetDroneDockerExecCmd())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logrus.Fatal(err)
	}
}

func getOauthToken(data []byte) (s string) {
	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}
	ctx := context.Background()
	credentials, err := google.CredentialsFromJSON(ctx, data, scopes...)
	if err == nil {
		token, err := credentials.TokenSource.Token()
		if err == nil {
			return token.AccessToken
		}
	}
	return
}

func setUsernameAndPassword(user string, pass string, is_wi bool) (u string, p string) {
	// decode the token if base64 encoded
	decoded, err := base64.StdEncoding.DecodeString(pass)
	if err == nil {
		pass = string(decoded)
	}
	// get oauth token and set username if using workload identity
	if is_wi {
		data := []byte(pass)
		pass = getOauthToken(data)
		user = "oauth2accesstoken"
	}
	return user, pass
}

func parseBoolOrDefault(defaultValue bool, s string) (result bool) {
	var err error
	result, err = strconv.ParseBool(s)
	if err != nil {
		result = defaultValue
	}

	return
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
