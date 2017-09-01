package main

import (
	"encoding/base64"
	"os"
	"os/exec"
	"strings"
)

const (
	DOCKER_USER    = "AWS"
	DEFAULT_REGION = "us-east-1"
)

func main() {
	ecrRegion := DEFAULT_REGION
	authorizationOptions := []string{
		"ecr",
		"get-authorization-token",
		"--output",
		"text",
	}

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
	if registryIds := getenv("PLUGIN_REGISTRY_IDS"); registryIds != "" {
		authorizationOptions = append(authorizationOptions, "--registry-ids")
		authorizationOptions = append(authorizationOptions, registryIds)
	}

	// Get credentials
	awsCli := exec.Command("aws", authorizationOptions...)
	output, err := awsCli.Output()
	if err != nil {
		os.Exit(1)
	}

	tokens := strings.Split(string(output), "\t")
	dockerPassword := strings.Split(decodeBase64(tokens[1]), ":")[1]
	dockerRegistry := tokens[3]

	os.Setenv("DOCKER_USERNAME", DOCKER_USER)
	os.Setenv("DOCKER_PASSWORD", dockerPassword)
	os.Setenv("DOCKER_REGISTRY", dockerRegistry)

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
