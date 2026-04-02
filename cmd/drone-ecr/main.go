package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/joho/godotenv"

	docker "github.com/drone-plugins/drone-docker"
)

const defaultRegion = "us-east-1"

func main() {
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	var (
		repo                = getenv("PLUGIN_REPO")
		registry            = getenv("PLUGIN_REGISTRY")
		region              = getenv("PLUGIN_REGION", "ECR_REGION", "AWS_REGION")
		key                 = getenv("PLUGIN_ACCESS_KEY", "ECR_ACCESS_KEY", "AWS_ACCESS_KEY_ID")
		secret              = getenv("PLUGIN_SECRET_KEY", "ECR_SECRET_KEY", "AWS_SECRET_ACCESS_KEY")
		create              = parseBoolOrDefault(false, getenv("PLUGIN_CREATE_REPOSITORY", "ECR_CREATE_REPOSITORY"))
		lifecyclePolicy     = getenv("PLUGIN_LIFECYCLE_POLICY")
		repositoryPolicy    = getenv("PLUGIN_REPOSITORY_POLICY")
		assumeRole          = getenv("PLUGIN_ASSUME_ROLE")
		externalId          = getenv("PLUGIN_EXTERNAL_ID")
		scanOnPush          = parseBoolOrDefault(false, getenv("PLUGIN_SCAN_ON_PUSH"))
		idToken             = os.Getenv("PLUGIN_OIDC_TOKEN_ID")
		skipPushIfTagExists = parseBoolOrDefault(false, getenv("PLUGIN_SKIP_PUSH_IF_TAG_EXISTS"))
	)

	if region == "" {
		region = defaultRegion
	}

	os.Setenv("AWS_REGION", region)

	if key != "" && secret != "" {
		os.Setenv("AWS_ACCESS_KEY_ID", key)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secret)
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatal(fmt.Sprintf("error creating aws config: %v", err))
	}

	svc := getECRClient(cfg, assumeRole, externalId, idToken)
	username, password, defaultRegistry, err := getAuthInfo(ctx, svc)

	if registry == "" {
		registry = defaultRegistry
	}

	if err != nil {
		log.Fatal(fmt.Sprintf("error getting ECR auth: %v", err))
	}

	if !strings.HasPrefix(repo, registry) {
		repo = fmt.Sprintf("%s/%s", registry, repo)
	}

	if create {
		err = ensureRepoExists(ctx, svc, trimHostname(repo, registry), scanOnPush)
		if err != nil {
			log.Fatal(fmt.Sprintf("error creating ECR repo: %v", err))
		}
		err = updateImageScanningConfig(ctx, svc, trimHostname(repo, registry), scanOnPush)
		if err != nil {
			log.Fatal(fmt.Sprintf("error updating scan on push for ECR repo: %v", err))
		}
	}

	if lifecyclePolicy != "" {
		p, err := os.ReadFile(lifecyclePolicy)
		if err != nil {
			log.Fatal(err)
		}
		if err := uploadLifeCyclePolicy(ctx, svc, string(p), trimHostname(repo, registry)); err != nil {
			log.Fatal(fmt.Sprintf("error uploading ECR lifecycle policy: %v", err))
		}
	}

	if repositoryPolicy != "" {
		p, err := os.ReadFile(repositoryPolicy)
		if err != nil {
			log.Fatal(err)
		}
		if err := uploadRepositoryPolicy(ctx, svc, string(p), trimHostname(repo, registry)); err != nil {
			log.Fatal(fmt.Sprintf("error uploading ECR repository policy. %v", err))
		}
	}

	os.Setenv("PLUGIN_REPO", repo)
	os.Setenv("PLUGIN_REGISTRY", registry)
	os.Setenv("DOCKER_USERNAME", username)
	os.Setenv("DOCKER_PASSWORD", password)
	os.Setenv("PLUGIN_REGISTRY_TYPE", "ECR")

	if skipPushIfTagExists {
		tagInput := getenv("PLUGIN_TAG", "PLUGIN_TAGS")
		var tags []string
		if tagInput == "" {
			tags = []string{"latest"}
		} else {
			for _, t := range strings.Split(tagInput, ",") {
				trimmed := strings.TrimSpace(t)
				if trimmed != "" {
					tags = append(tags, trimmed)
				}
			}
		}

	repositoryName := trimHostname(repo, registry)
	for _, t := range tags {
		exists, err := tagExists(ctx, svc, repositoryName, t)
		if err != nil {
			slog.Error("error checking if image exists for tag", "tag", t, "error", err)
			os.Exit(1)
		}
		if exists {
			slog.Info("image tag exists, skipping push", "repo", repo, "tag", t)
			os.Exit(0)
		}
	}
	}

	cmd := exec.Command(docker.GetDroneDockerExecCmd())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

func trimHostname(repo, registry string) string {
	repo = strings.TrimPrefix(repo, registry)
	repo = strings.TrimLeft(repo, "/")
	return repo
}

func ensureRepoExists(ctx context.Context, svc *ecr.Client, name string, scanOnPush bool) error {
	_, err := svc.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(name),
		ImageScanningConfiguration: &ecrtypes.ImageScanningConfiguration{
			ScanOnPush: scanOnPush,
		},
	})
	if err != nil {
		var rae *ecrtypes.RepositoryAlreadyExistsException
		if errors.As(err, &rae) {
			return nil
		}
		return err
	}
	return nil
}

func updateImageScanningConfig(ctx context.Context, svc *ecr.Client, name string, scanOnPush bool) error {
	_, err := svc.PutImageScanningConfiguration(ctx, &ecr.PutImageScanningConfigurationInput{
		RepositoryName: aws.String(name),
		ImageScanningConfiguration: &ecrtypes.ImageScanningConfiguration{
			ScanOnPush: scanOnPush,
		},
	})
	return err
}

func uploadLifeCyclePolicy(ctx context.Context, svc *ecr.Client, lifecyclePolicy string, name string) error {
	_, err := svc.PutLifecyclePolicy(ctx, &ecr.PutLifecyclePolicyInput{
		LifecyclePolicyText: aws.String(lifecyclePolicy),
		RepositoryName:      aws.String(name),
	})
	return err
}

func uploadRepositoryPolicy(ctx context.Context, svc *ecr.Client, repositoryPolicy string, name string) error {
	_, err := svc.SetRepositoryPolicy(ctx, &ecr.SetRepositoryPolicyInput{
		PolicyText:     aws.String(repositoryPolicy),
		RepositoryName: aws.String(name),
	})
	return err
}

func getAuthInfo(ctx context.Context, svc *ecr.Client) (username, password, registry string, err error) {
	var result *ecr.GetAuthorizationTokenOutput
	var decoded []byte

	result, err = svc.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return
	}

	if len(result.AuthorizationData) == 0 {
		err = fmt.Errorf("no authorization data returned from ECR")
		return
	}

	auth := result.AuthorizationData[0]
	token := *auth.AuthorizationToken
	decoded, err = base64.StdEncoding.DecodeString(token)
	if err != nil {
		return
	}

	registry = strings.TrimPrefix(*auth.ProxyEndpoint, "https://")
	creds := strings.SplitN(string(decoded), ":", 2)
	if len(creds) < 2 {
		err = fmt.Errorf("invalid ECR authorization token format")
		return
	}
	username = creds[0]
	password = creds[1]
	return
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

func getECRClient(cfg aws.Config, role string, externalId string, idToken string) *ecr.Client {
	if role == "" {
		return ecr.NewFromConfig(cfg)
	}

	stsSvc := sts.NewFromConfig(cfg)

	if idToken != "" {
		provider := stscreds.NewWebIdentityRoleProvider(stsSvc, role, identityToken(idToken))
		cfg.Credentials = aws.NewCredentialsCache(provider)
		return ecr.NewFromConfig(cfg)
	}

	var provider *stscreds.AssumeRoleProvider
	if externalId != "" {
		provider = stscreds.NewAssumeRoleProvider(stsSvc, role, func(o *stscreds.AssumeRoleOptions) {
			o.ExternalID = &externalId
		})
	} else {
		provider = stscreds.NewAssumeRoleProvider(stsSvc, role)
	}
	cfg.Credentials = aws.NewCredentialsCache(provider)
	return ecr.NewFromConfig(cfg)
}

func tagExists(ctx context.Context, svc *ecr.Client, repository, tag string) (bool, error) {
	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repository),
		ImageIds: []ecrtypes.ImageIdentifier{
			{ImageTag: aws.String(tag)},
		},
	}
	output, err := svc.DescribeImages(ctx, input)
	if err != nil {
		var inf *ecrtypes.ImageNotFoundException
		if errors.As(err, &inf) {
			return false, nil
		}
		return false, err
	}
	return len(output.ImageDetails) > 0, nil
}

type identityToken string

func (t identityToken) GetIdentityToken() ([]byte, error) {
	return []byte(t), nil
}
