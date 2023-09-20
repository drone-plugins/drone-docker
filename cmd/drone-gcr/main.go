package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
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

type Config struct {
	Repo             string
	Registry         string
	Password         string
	WorkloadIdentity bool
	Username         string
	RegistryType     string
}

func loadConfig() Config {
	// Default username
	username := "_json_key"

	// Load env-file if it exists
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		if err := godotenv.Load(env); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	location := getenv("PLUGIN_LOCATION")
	repo := getenv("PLUGIN_REPO")

	password := getenv(
		"PLUGIN_JSON_KEY",
		"GCR_JSON_KEY",
		"GOOGLE_CREDENTIALS",
		"TOKEN",
	)
	workloadIdentity := parseBoolOrDefault(false, getenv("PLUGIN_WORKLOAD_IDENTITY"))
	username, password = setUsernameAndPassword(username, password, workloadIdentity)

	registryType := getenv("PLUGIN_REGISTRY_TYPE")
	if registryType == "" {
		registryType = "GCR"
	}

	registry := getenv("PLUGIN_REGISTRY")
	if registry == "" {
		switch registryType {
		case "GCR":
			registry = "gcr.io"
		case "GAR":
			if location == "" {
				logrus.Fatalf("Error: For REGISTRY_TYPE of GAR, LOCATION must be set")
			}
			registry = fmt.Sprintf("%s-docker.pkg.dev", location)
		default:
			logrus.Fatalf("Unsupported registry type: %s", registryType)
		}
	}

	if !strings.HasPrefix(repo, registry) {
		repo = path.Join(registry, repo)
	}

	return Config{
		Repo:             repo,
		Registry:         registry,
		Password:         password,
		WorkloadIdentity: workloadIdentity,
		Username:         username,
		RegistryType:     registryType,
	}
}

func main() {
	config := loadConfig()

	os.Setenv("PLUGIN_REPO", config.Repo)
	os.Setenv("PLUGIN_REGISTRY", config.Registry)
	os.Setenv("DOCKER_USERNAME", config.Username)
	os.Setenv("DOCKER_PASSWORD", config.Password)
	os.Setenv("PLUGIN_REGISTRY_TYPE", config.RegistryType)

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

func setUsernameAndPassword(user string, pass string, workloadIdentity bool) (u string, p string) {
	// decode the token if base64 encoded
	decoded, err := base64.StdEncoding.DecodeString(pass)
	if err == nil {
		pass = string(decoded)
	}
	// get oauth token and set username if using workload identity
	if workloadIdentity {
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
