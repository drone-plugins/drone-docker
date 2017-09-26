package main

import (
	"encoding/base64"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

var (
	repo      = getenv("PLUGIN_REPO")
	registry  = getenv("PLUGIN_REGISTRY")
	region    = getenv("PLUGIN_REGION", "ECR_REGION")
	accessKey = getenv("PLUGIN_ACCESS_KEY", "ECR_ACCESS_KEY")
	secretKey = getenv("PLUGIN_SECRET_KEY", "ECR_SECRET_KEY")
)

func main() {
	// default region value
	if region == "" {
		region = "us-east-1"
	}
	// ensure aws env vars
	if accessKey == "" || secretKey == "" {
		os.Exit(1)
	}

	os.Setenv("AWS_ACCESS_KEY_ID", accessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)

	sess := session.New()
	svc := ecr.New(sess, aws.NewConfig().WithMaxRetries(10).WithRegion(region))

	resp, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		os.Exit(1)
	}

	auth := resp.AuthorizationData[0]
	data, err := base64.StdEncoding.DecodeString(*auth.AuthorizationToken)
	if err != nil {
		os.Exit(1)
	}

	token := strings.SplitN(string(data), ":", 2)

	os.Setenv("PLUGIN_USERNAME", token[0])
	os.Setenv("PLUGIN_PASSWORD", token[1])

	os.Setenv("PLUGIN_REPO", repo)
	os.Setenv("PLUGIN_REGISTRY", *auth.ProxyEndpoint)

	// invoke the base docker plugin binary
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
