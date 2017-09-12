package main

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"os"
	"os/exec"
)

const (
	DockerUser    = "AWS"
	DefaultRegion = "us-east-1"
)

func main() {
	var registryIds []*string
	var (
		ecrRegion  = getenv("ECR_REGION", "PLUGIN_REGION")
		registries = getenv("ECR_REGISTRY_IDS", "PLUGIN_REGISTRY_IDS")
		accessKey  = getenv(
			"AWS_ACCESS_KEY_ID",
			"AWS_ACCESS_KEY",
			"ECR_ACCESS_KEY",
			"PLUGIN_ACCESS_KEY",
		)
		secretKey = getenv(
			"AWS_SECRET_ACCESS_KEY",
			"AWS_SECRET_KEY",
			"ECR_SECRET_KEY",
			"PLUGIN_SECRET_KEY",
		)
	)

	if ecrRegion == "" {
		ecrRegion = DefaultRegion
	}

	// Useful when using a registry from another account
	if registries != "" {
		registryIds = append(registryIds, &registries)
	}

	password, registry, err := getCredentials(ecrRegion, accessKey, secretKey, registryIds)
	if err != nil {
		fmt.Println(err)
		return
	}

	os.Setenv("DOCKER_USERNAME", DockerUser)
	os.Setenv("DOCKER_PASSWORD", password)
	os.Setenv("DOCKER_REGISTRY", registry)

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

func decodeBase64(data string) string {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func getECRClient(region string, accessKey string, secretKey string) (*ecr.ECR, error) {
	config := aws.Config{
		Region: aws.String(region),
	}
	if accessKey != "" && secretKey != "" {
		creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
		config.Credentials = creds
	}

	sess, err := session.NewSession(&config)
	if err != nil {
		return nil, err
	}
	return ecr.New(sess), nil
}

func getCredentials(region string, accessKey string, secretKey string, registryIds []*string) (string, string, error) {
	client, err := getECRClient(region, accessKey, secretKey)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: registryIds,
	}
	result, err := client.GetAuthorizationToken(input)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	// Password has a prefix "AWS:" which is not necessary
	password := decodeBase64(*result.AuthorizationData[0].AuthorizationToken)[4:]
	registry := *result.AuthorizationData[0].ProxyEndpoint
	return password, registry, nil
}
