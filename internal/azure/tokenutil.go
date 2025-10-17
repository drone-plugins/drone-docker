package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const DefaultResource = "https://management.azure.com/"

// GetAADAccessTokenViaClientAssertion exchanges an external OIDC ID token for an Azure AD access token

func GetAADAccessTokenViaClientAssertion(ctx context.Context, tenantID, clientID, oidcToken, resource string) (string, error) {
	if resource == "" {
		resource = DefaultResource
	}

	form := url.Values{
		"client_id":             {clientID},
		"scope":                 {resource + ".default"},
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {oidcToken},
	}
	endpoint := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
	resp, err := http.PostForm(endpoint, form)
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
