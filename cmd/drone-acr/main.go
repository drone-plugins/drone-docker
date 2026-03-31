package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/joho/godotenv"

	docker "github.com/drone-plugins/drone-docker"
	azureutil "github.com/drone-plugins/drone-docker/internal/azure"
)

type subscriptionUrlResponse struct {
	Value []struct {
		ID string `json:"id"`
	} `json:"value"`
}

const (
	acrCertFile              = "acr-cert.pem"
	azSubscriptionApiVersion = "2021-04-01"
	azSubscriptionBaseUrl    = "https://management.azure.com/subscriptions/"
	basePublicUrl            = "https://portal.azure.com/#view/Microsoft_Azure_ContainerRegistries/TagMetadataBlade/registryId/"
	defaultUsername          = "00000000-0000-0000-0000-000000000000"

	// Environment variable names for Azure Environment Credential
	clientIdEnv        = "AZURE_CLIENT_ID"
	clientSecretKeyEnv = "AZURE_CLIENT_SECRET"
	tenantKeyEnv       = "AZURE_TENANT_ID"
	certPathEnv        = "AZURE_CLIENT_CERTIFICATE_PATH"
)

var (
	acrCertPath = filepath.Join(os.TempDir(), acrCertFile)
)

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	var (
		repo     = getenv("PLUGIN_REPO")
		registry = getenv("PLUGIN_REGISTRY")

		// If these credentials are provided, they will be directly used
		// for docker login
		username = getenv("SERVICE_PRINCIPAL_CLIENT_ID")
		password = getenv("SERVICE_PRINCIPAL_CLIENT_SECRET")

		// Service principal credentials
		clientId       = getenv("CLIENT_ID", "AZURE_CLIENT_ID", "AZURE_APP_ID", "PLUGIN_CLIENT_ID")
		clientSecret   = getenv("CLIENT_SECRET", "PLUGIN_CLIENT_SECRET")
		clientCert     = getenv("CLIENT_CERTIFICATE", "PLUGIN_CLIENT_CERTIFICATE")
		tenantId       = getenv("TENANT_ID", "AZURE_TENANT_ID", "PLUGIN_TENANT_ID")
		subscriptionId = getenv("SUBSCRIPTION_ID", "PLUGIN_SUBSCRIPTION_ID")
		publicUrl      = getenv("DAEMON_REGISTRY", "PLUGIN_DAEMON_REGISTRY")
		authorityHost  = getenv("AZURE_AUTHORITY_HOST", "PLUGIN_AZURE_AUTHORITY_HOST")
		idToken        = getenv("PLUGIN_OIDC_TOKEN_ID")
	)

	// default registry value
	if registry == "" {
		registry = "azurecr.io"
	}

	// Get auth if username and password is not specified
	if username == "" && password == "" {
		// docker login credentials are not provided
		var err error
	username = defaultUsername
	if idToken != "" && clientId != "" && tenantId != "" {
		slog.Debug("using OIDC authentication flow")
		var aadToken string
		aadToken, err = azureutil.GetAADAccessTokenViaClientAssertion(context.Background(), tenantId, clientId, idToken, authorityHost)
		if err != nil {
			slog.Error("failed to get AAD access token", "error", err)
			os.Exit(1)
		}
			var p string
			p, err = getPublicUrl(aadToken, registry, subscriptionId)
			if err == nil {
				publicUrl = p
			} else {
				fmt.Fprintf(os.Stderr, "failed to get public url with error: %s\n", err)
			}
		password, err = fetchACRToken(tenantId, aadToken, registry)
		if err != nil {
			slog.Error("failed to fetch ACR token", "error", err)
			os.Exit(1)
		}
	} else {
		password, publicUrl, err = getAuth(clientId, clientSecret, clientCert, tenantId, subscriptionId, registry)
		if err != nil {
			slog.Error("failed to get auth", "error", err)
			os.Exit(1)
		}
	}
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
	os.Setenv("PLUGIN_REGISTRY_TYPE", "ACR")
	if publicUrl != "" {
		// Set this env variable if public URL for artifact is available
		// If not, we will fall back to registry url
		os.Setenv("ARTIFACT_REGISTRY", publicUrl)
	}

	// invoke the base docker plugin binary
	cmd := exec.Command(docker.GetDroneDockerExecCmd())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

func getAuth(clientId, clientSecret, clientCert, tenantId, subscriptionId, registry string) (string, string, error) {
	// Verify inputs
	if tenantId == "" {
		return "", "", fmt.Errorf("tenantId cannot be empty for AAD authentication")
	}
	if clientId == "" {
		return "", "", fmt.Errorf("clientId cannot be empty for AAD authentication")
	}
	if clientSecret == "" && clientCert == "" {
		return "", "", fmt.Errorf("one of client secret or client cert should be defined")
	}

	// Setup cert
	if clientCert != "" {
		err := setupACRCert(clientCert, acrCertPath)
		if err != nil {
			fmt.Errorf("failed to push setup cert file: %w", err)
		}
	}

	// Get AZ env
	if err := os.Setenv(clientIdEnv, clientId); err != nil {
		return "", "", fmt.Errorf("failed to set env variable client Id: %w", err)
	}
	if err := os.Setenv(clientSecretKeyEnv, clientSecret); err != nil {
		return "", "", fmt.Errorf("failed to set env variable client secret: %w", err)
	}
	if err := os.Setenv(tenantKeyEnv, tenantId); err != nil {
		return "", "", fmt.Errorf("failed to set env variable tenant Id: %w", err)
	}
	if err := os.Setenv(certPathEnv, acrCertPath); err != nil {
		return "", "", fmt.Errorf("failed to set env variable cert path: %w", err)
	}
	env, err := azidentity.NewEnvironmentCredential(nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to get env credentials from azure: %w", err)
	}
	os.Unsetenv(clientIdEnv)
	os.Unsetenv(clientSecretKeyEnv)
	os.Unsetenv(tenantKeyEnv)
	os.Unsetenv(certPathEnv)

	// Fetch AAD token
	policy := policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	}
	aadToken, err := env.GetToken(context.Background(), policy)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch access token: %w", err)
	}

	// Get public URL for artifacts
	publicUrl, err := getPublicUrl(aadToken.Token, registry, subscriptionId)
	if err != nil {
		// execution should not fail because of this error
		fmt.Fprintf(os.Stderr, "failed to get public url with error: %s\n", err)
	}

	// Fetch token
	ACRToken, err := fetchACRToken(tenantId, aadToken.Token, registry)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch ACR token: %w", err)
	}
	return ACRToken, publicUrl, nil
}

func fetchACRToken(tenantId, token, registry string) (string, error) {
	// oauth exchange
	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {registry},
		"tenant":       {tenantId},
		"access_token": {token},
	}
	jsonResponse, err := http.PostForm(fmt.Sprintf("https://%s/oauth2/exchange", registry), formData)
	if err != nil || jsonResponse == nil {
		return "", fmt.Errorf("failed to fetch ACR token: %w", err)
	}

	// fetch token from response
	var response map[string]interface{}
	err = json.NewDecoder(jsonResponse.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode oauth exchange response: %w", err)
	}

	// Parse the refresh_token from the response
	if t, found := response["refresh_token"]; found {
		if refreshToken, ok := t.(string); ok {
			return refreshToken, nil
		}
		return "", fmt.Errorf("failed to cast refresh token from acr")
	}
	return "", fmt.Errorf("refresh token not found in response of oauth exchange call: %w", err)
}

func setupACRCert(cert, certPath string) error {
	decoded, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return fmt.Errorf("failed to base64 decode ACR certificate: %w", err)
	}
	err = ioutil.WriteFile(certPath, decoded, 0644)
	if err != nil {
		return fmt.Errorf("failed to write ACR certificate: %w", err)
	}
	return nil
}

func getPublicUrl(token, registryUrl, subscriptionId string) (string, error) {
	if len(subscriptionId) == 0 || registryUrl == "" {
		return "", nil
	}

	registry := strings.Split(registryUrl, ".")[0]
	filter := fmt.Sprintf("resourceType eq 'Microsoft.ContainerRegistry/registries' and name eq '%s'", registry)
	params := url.Values{}
	params.Add("$filter", filter)
	params.Add("api-version", azSubscriptionApiVersion)
	params.Add("$select", "id")
	url := azSubscriptionBaseUrl + subscriptionId + "/resources?" + params.Encode()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("failed to create request for getting container registry setting: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("failed to send request for getting container registry setting: %w", err)
	}
	defer res.Body.Close()

	var response subscriptionUrlResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to send request for getting container registry setting: %w", err)
	}
	if len(response.Value) == 0 {
		return "", fmt.Errorf("no id present for base url")
	}
	return basePublicUrl + encodeParam(response.Value[0].ID), nil
}

func encodeParam(s string) string {
	return url.QueryEscape(s)
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
