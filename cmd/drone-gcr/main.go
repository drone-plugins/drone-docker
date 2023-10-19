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

	docker "github.com/drone-plugins/drone-docker"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iamcredentials/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sts/v1"
)

type Config struct {
	Repo             string
	Registry         string
	Password         string
	WorkloadIdentity bool
	Username         string
	RegistryType     string
	AccessToken      string
}

type staticTokenSource struct {
	token *oauth2.Token
}

func (s *staticTokenSource) Token() (*oauth2.Token, error) {
	return s.token, nil
}

func loadConfig() Config {
	// Default username
	username := "_json_key"
	var config Config

	// Load env-file if it exists
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		if err := godotenv.Load(env); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	idToken := getenv("PLUGIN_OIDC_TOKEN_ID")
	projectId := getenv("PLUGIN_PROJECT_NUMBER")
	poolId := getenv("PLUGIN_POOL_ID")
	providerId := getenv("PLUGIN_PROVIDER_ID")
	serviceAccountEmail := getenv("PLUGIN_SERVICE_ACCOUNT_EMAIL")

	if idToken != "" && projectId != "" && poolId != "" && providerId != "" && serviceAccountEmail != "" {
		federalToken, err := getFederalToken(idToken, projectId, poolId, providerId)
		if err != nil {
			logrus.Fatalf("Error (getFederalToken): %s", err)
		}
		accessToken, err := getGoogleCloudAccessToken(federalToken, serviceAccountEmail)
		if err != nil {
			logrus.Fatalf("Error (getGoogleCloudAccessToken): %s", err)
		}
		config.AccessToken = accessToken
	} else {
		password := getenv(
			"PLUGIN_JSON_KEY",
			"GCR_JSON_KEY",
			"GOOGLE_CREDENTIALS",
			"TOKEN",
		)
		config.WorkloadIdentity = parseBoolOrDefault(false, getenv("PLUGIN_WORKLOAD_IDENTITY"))
		config.Username, config.Password = setUsernameAndPassword(username, password, config.WorkloadIdentity)
	}

	location := getenv("PLUGIN_LOCATION")
	repo := getenv("PLUGIN_REPO")

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
	config.Repo = repo
	config.Registry = registry
	config.RegistryType = registryType
	return config
}

func main() {
	config := loadConfig()
	if config.AccessToken != "" {
		os.Setenv("ACCESS_TOKEN", config.AccessToken)
	} else if config.Username != "" && config.Password != "" {
		os.Setenv("DOCKER_USERNAME", config.Username)
		os.Setenv("DOCKER_PASSWORD", config.Password)
		os.Setenv("", strconv.FormatBool(config.WorkloadIdentity))
	}

	os.Setenv("PLUGIN_REPO", config.Repo)
	os.Setenv("PLUGIN_REGISTRY", config.Registry)
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

func getFederalToken(idToken, projectNumber, poolId, providerId string) (string, error) {
	ctx := context.Background()
	stsService, err := sts.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		return "", err
	}
	audience := fmt.Sprintf("//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s", projectNumber, poolId, providerId)
	tokenRequest := &sts.GoogleIdentityStsV1ExchangeTokenRequest{
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		SubjectToken:       idToken,
		Audience:           audience,
		Scope:              "https://www.googleapis.com/auth/cloud-platform",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:id_token",
	}
	tokenResponse, err := stsService.V1.Token(tokenRequest).Do()
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func getGoogleCloudAccessToken(federatedToken string, serviceAccountEmail string) (string, error) {
	ctx := context.Background()
	tokenSource := &staticTokenSource{
		token: &oauth2.Token{AccessToken: federatedToken},
	}
	service, err := iamcredentials.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return "", err
	}

	name := "projects/-/serviceAccounts/" + serviceAccountEmail
	rb := &iamcredentials.GenerateAccessTokenRequest{
		Scope: []string{"https://www.googleapis.com/auth/cloud-platform"},
	}

	resp, err := service.Projects.ServiceAccounts.GenerateAccessToken(name, rb).Do()
	if err != nil {
		return "", err
	}

	return resp.AccessToken, nil
}
