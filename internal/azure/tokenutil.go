package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultResource = "https://management.azure.com/"
const defaultAuthorityHost = "https://login.microsoftonline.com"
const defaultHTTPTimeout = 30 * time.Second

// GetAADAccessTokenViaClientAssertion exchanges an external OIDC ID token for an Azure AD access token

func GetAADAccessTokenViaClientAssertion(ctx context.Context, tenantID, clientID, oidcToken, authorityHost string) (string, error) {
    resource := DefaultResource

	form := url.Values{
		"client_id":             {clientID},
		"scope":                 {resource + ".default"},
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {oidcToken},
	}

	base := authorityHost
	if strings.TrimSpace(base) == "" {
		base = defaultAuthorityHost
	}
	base = strings.TrimRight(base, "/")
	endpoint := fmt.Sprintf("%s/%s/oauth2/v2.0/token", base, tenantID)

	client := &http.Client{Timeout: defaultHTTPTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var aadErr struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		limited := io.LimitedReader{R: resp.Body, N: 4096}
		_ = json.NewDecoder(&limited).Decode(&aadErr)
		if aadErr.Error != "" {
			return "", fmt.Errorf("AAD token request failed: status=%d, error=%s", resp.StatusCode, aadErr.Error)
		}
		return "", fmt.Errorf("AAD token request failed: status=%d", resp.StatusCode)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("AAD token response missing access_token")
	}
	return payload.AccessToken, nil
}
