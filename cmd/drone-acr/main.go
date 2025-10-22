package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

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
		clientSecret   = getenv("CLIENT_SECRET")
		clientCert     = getenv("CLIENT_CERTIFICATE")
		tenantId       = getenv("TENANT_ID", "AZURE_TENANT_ID", "PLUGIN_TENANT_ID")
		subscriptionId = getenv("SUBSCRIPTION_ID")
		publicUrl      = getenv("DAEMON_REGISTRY")
		authorityHost  = getenv("AZURE_AUTHORITY_HOST")
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
			var aadToken string
			aadToken, err = azureutil.GetAADAccessTokenViaClientAssertion(context.Background(), tenantId, clientId, idToken, authorityHost)
			if err != nil {
				logrus.Fatal(err)
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
				logrus.Fatal(err)
			}
		} else {
			password, publicUrl, err = getAuth(clientId, clientSecret, clientCert, tenantId, subscriptionId, registry)
			if err != nil {
				logrus.Fatal(err)
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
		logrus.Fatal(err)
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
			errors.Wrap(err, "failed to push setup cert file")
		}
	}

	// Get AZ env
	if err := os.Setenv(clientIdEnv, clientId); err != nil {
		return "", "", errors.Wrap(err, "failed to set env variable client Id")
	}
	if err := os.Setenv(clientSecretKeyEnv, clientSecret); err != nil {
		return "", "", errors.Wrap(err, "failed to set env variable client secret")
	}
	if err := os.Setenv(tenantKeyEnv, tenantId); err != nil {
		return "", "", errors.Wrap(err, "failed to set env variable tenant Id")
	}
	if err := os.Setenv(certPathEnv, acrCertPath); err != nil {
		return "", "", errors.Wrap(err, "failed to set env variable cert path")
	}
	env, err := azidentity.NewEnvironmentCredential(nil)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get env credentials from azure")
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
		return "", "", errors.Wrap(err, "failed to fetch access token")
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
		return "", "", errors.Wrap(err, "failed to fetch ACR token")
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
		return "", errors.Wrap(err, "failed to fetch ACR token")
	}

	// fetch token from response
	var response map[string]interface{}
	err = json.NewDecoder(jsonResponse.Body).Decode(&response)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode oauth exchange response")
	}

	// Parse the refresh_token from the response
	if t, found := response["refresh_token"]; found {
		if refreshToken, ok := t.(string); ok {
			return refreshToken, nil
		}
		return "", errors.New("failed to cast refresh token from acr")
	}
	return "", errors.Wrap(err, "refresh token not found in response of oauth exchange call")
}

func setupACRCert(cert, certPath string) error {
	decoded, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return errors.Wrap(err, "failed to base64 decode ACR certificate")
	}
	err = ioutil.WriteFile(certPath, decoded, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write ACR certificate")
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
		return "", errors.Wrap(err, "failed to create request for getting container registry setting")
	}

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", errors.Wrap(err, "failed to send request for getting container registry setting")
	}
	defer res.Body.Close()

	var response subscriptionUrlResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request for getting container registry setting")
	}
	if len(response.Value) == 0 {
		return "", errors.New("no id present for base url")
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
