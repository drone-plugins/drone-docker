package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	docker "github.com/drone-plugins/drone-docker"
)

type ecrAPI interface {
	DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error)
}

const defaultRegion = "us-east-1"

func main() {
	// Load env-file if it exists first
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

	// set the region
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

	svc := getECRClient(ctx, cfg, assumeRole, externalId, idToken)
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
		err = updateImageScannningConfig(ctx, svc, trimHostname(repo, registry), scanOnPush)
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

	// Skip if tag already exits for both mutable and immutable repos
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
				logrus.Fatalf("Error checking if image exists for tag %s: %v", t, err)
			}
			if exists {
				logrus.Infof("%s:%s: Image tag exists. Skipping push.", repo, t)
				os.Exit(0)
			}
		}
	}

	// invoke the base docker plugin binary
	cmd := exec.Command(docker.GetDroneDockerExecCmd())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		logrus.Fatal(err)
	}
}

func trimHostname(repo, registry string) string {
	repo = strings.TrimPrefix(repo, registry)
	repo = strings.TrimLeft(repo, "/")
	return repo
}

func ensureRepoExists(ctx context.Context, svc *ecr.Client, name string, scanOnPush bool) (err error) {
	input := &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(name),
		ImageScanningConfiguration: &types.ImageScanningConfiguration{
			ScanOnPush: scanOnPush,
		},
	}
	_, err = svc.CreateRepository(ctx, input)
	if err != nil {
		// Check if repository already exists
		var repoExistsErr *types.RepositoryAlreadyExistsException
		if isErrorType(err, &repoExistsErr) {
			// eat it, we skip checking for existing to save two requests
			err = nil
		}
	}

	return
}

func isErrorType[T error](err error, target *T) bool {
	if err == nil {
		return false
	}
	for err != nil {
		if matched, ok := err.(T); ok {
			*target = matched
			return true
		}
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return false
}

func updateImageScannningConfig(ctx context.Context, svc *ecr.Client, name string, scanOnPush bool) (err error) {
	input := &ecr.PutImageScanningConfigurationInput{
		RepositoryName: aws.String(name),
		ImageScanningConfiguration: &types.ImageScanningConfiguration{
			ScanOnPush: scanOnPush,
		},
	}
	_, err = svc.PutImageScanningConfiguration(ctx, input)

	return err
}

func uploadLifeCyclePolicy(ctx context.Context, svc *ecr.Client, lifecyclePolicy string, name string) (err error) {
	input := &ecr.PutLifecyclePolicyInput{
		LifecyclePolicyText: aws.String(lifecyclePolicy),
		RepositoryName:      aws.String(name),
	}
	_, err = svc.PutLifecyclePolicy(ctx, input)

	return err
}

func uploadRepositoryPolicy(ctx context.Context, svc *ecr.Client, repositoryPolicy string, name string) (err error) {
	input := &ecr.SetRepositoryPolicyInput{
		PolicyText:     aws.String(repositoryPolicy),
		RepositoryName: aws.String(name),
	}
	_, err = svc.SetRepositoryPolicy(ctx, input)

	return err
}

func getAuthInfo(ctx context.Context, svc *ecr.Client) (username, password, registry string, err error) {
	var result *ecr.GetAuthorizationTokenOutput
	var decoded []byte

	result, err = svc.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return
	}

	auth := result.AuthorizationData[0]
	token := *auth.AuthorizationToken
	decoded, err = base64.StdEncoding.DecodeString(token)
	if err != nil {
		return
	}

	registry = strings.TrimPrefix(*auth.ProxyEndpoint, "https://")
	creds := strings.Split(string(decoded), ":")
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

func getECRClient(ctx context.Context, cfg aws.Config, role string, externalId string, idToken string) *ecr.Client {
	if role == "" {
		return ecr.NewFromConfig(cfg)
	}

	stsClient := sts.NewFromConfig(cfg)

	if idToken != "" {
		tempFile, err := os.CreateTemp("/tmp", "idToken-*.jwt")
		if err != nil {
			log.Fatalf("Failed to create temporary file: %v", err)
		}
		defer tempFile.Close()

		if err := os.Chmod(tempFile.Name(), 0600); err != nil {
			log.Fatalf("Failed to set file permissions: %v", err)
		}

		if _, err := tempFile.WriteString(idToken); err != nil {
			log.Fatalf("Failed to write ID token to temporary file: %v", err)
		}

		// Create credentials using the path to the ID token file
		creds := stscreds.NewWebIdentityRoleProvider(stsClient, role, stscreds.IdentityTokenFile(tempFile.Name()))
		cfg.Credentials = aws.NewCredentialsCache(creds)
		return ecr.NewFromConfig(cfg)
	} else if externalId != "" {
		creds := stscreds.NewAssumeRoleProvider(stsClient, role, func(o *stscreds.AssumeRoleOptions) {
			o.ExternalID = &externalId
		})
		cfg.Credentials = aws.NewCredentialsCache(creds)
		return ecr.NewFromConfig(cfg)
	} else {
		creds := stscreds.NewAssumeRoleProvider(stsClient, role)
		cfg.Credentials = aws.NewCredentialsCache(creds)
		return ecr.NewFromConfig(cfg)
	}
}

func tagExists(ctx context.Context, svc ecrAPI, repository, tag string) (bool, error) {
	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repository),
		ImageIds: []types.ImageIdentifier{
			{ImageTag: aws.String(tag)},
		},
	}
	output, err := svc.DescribeImages(ctx, input)
	if err != nil {
		// Check if image not found
		var imageNotFoundErr *types.ImageNotFoundException
		if isErrorType(err, &imageNotFoundErr) {
			return false, nil
		}
		return false, err
	}
	return len(output.ImageDetails) > 0, nil
}
