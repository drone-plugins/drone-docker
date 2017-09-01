package main

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"os"
	"os/exec"
	"strings"
)

const (
	DOCKER_USER    = "AWS"
	DEFAULT_REGION = "us-east-1"
)

func main() {
	var registryIds []*string
	ecrRegion := DEFAULT_REGION

	// If environment variables are not specified
	// awscli will assume instance role
	if getenv("PLUGIN_ECR_REGION") != "" {
		ecrRegion = getenv("PLUGIN_ECR_REGION")
	}
	os.Setenv("AWS_DEFAULT_REGION", ecrRegion)

	if accessKey := getenv("PLUGIN_ACCESS_KEY"); accessKey != "" {
		os.Setenv("AWS_ACCESS_KEY_ID", accessKey)
	}

	if secretKey := getenv("PLUGIN_SECRET_KEY"); secretKey != "" {
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)
	}

	// Useful when using a registry from another account
	if registries := getenv("PLUGIN_REGISTRY_IDS"); registries != "" {
		registryIds = append(registryIds, &registries)
	}
	password, registry, err := getCredentials(ecrRegion, registryIds)
	if err != nil {
		fmt.Println(err)
		return
	}
	os.Setenv("DOCKER_USERNAME", DOCKER_USER)
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

func getRegistry(proxyEndpoint string) string {
	// proxyEndpoint has format
	// https://<registryid>.dkr.ecr.us-east-1.amazonaws.com
	return strings.Split(proxyEndpoint, ".")[0][8:]
}

func getECRClient(region string) (*ecr.ECR, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	return ecr.New(sess), nil
}

func getCredentials(region string, registryIds []*string) (string, string, error) {
	client, err := getECRClient(region)
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
	registry := getRegistry(*result.AuthorizationData[0].ProxyEndpoint)
	return password, registry, nil
}
