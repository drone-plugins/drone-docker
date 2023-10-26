package gcp

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/iamcredentials/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sts/v1"
)

type staticTokenSource struct {
	token *oauth2.Token
}

func (s *staticTokenSource) Token() (*oauth2.Token, error) {
	return s.token, nil
}

func GetFederalToken(idToken, projectNumber, poolId, providerId string) (string, error) {
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

func GetGoogleCloudAccessToken(federatedToken string, serviceAccountEmail string) (string, error) {
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
