# build once push everywhere (ecr/gcr/docker hub)

Fork of `drone-plugins/drone-docker`.

This uses the `config.json` file and these standalone docker credential helpers to authenticate with the image registries:

- https://github.com/GoogleCloudPlatform/docker-credential-gcr
- https://github.com/awslabs/amazon-ecr-credential-helper

The plugin image is at: `tonglil/drone-docker-multiple`

The repository must be marked as `trusted`, or the plugin image must be included in the `DRONE_RUNNER_PRIVILEGED_IMAGES` [configuration environment variable](https://docs.drone.io/runner/docker/configuration/reference/drone-runner-privileged-images/).

Example config:

```yaml
---
kind: pipeline
type: docker
name: default

# the values assume using the default /drone/src/ workspace

steps:
  - name: gcr_creds
    image: alpine
    commands:
      - mkdir -p $(dirname $GOOGLE_APPLICATION_CREDENTIALS)
      - printf "%s" "$GOOGLE_CREDENTIALS" > $GOOGLE_APPLICATION_CREDENTIALS
    environment:
      GOOGLE_APPLICATION_CREDENTIALS: /drone/src/.config/gcloud/application_default_credentials.json
      GOOGLE_CREDENTIALS:
        from_secret: google_credentials_gke

  - name: ecr_creds
    image: alpine
    commands:
      - mkdir -p $(dirname $AWS_CREDENTIALS)
      - printf "%s" "$SHARED_CREDENTIALS" > $AWS_CREDENTIALS
    environment:
      AWS_CREDENTIALS: /drone/src/.aws/credentials
      SHARED_CREDENTIALS:
        from_secret: aws_shared_credentials

  - name: docker
    image: tonglil/drone-docker-multiple
    privileged: true # this can be removed if the plugin is included in `DRONE_RUNNER_PRIVILEGED_IMAGES`
    settings:
      config:
        from_secret: docker_cred_helpers_config
      repos:
        - tonglil/example-image
        - us.gcr.io/my-project/example-image
        - 123456789012.dkr.ecr.us-west-2.amazonaws.com/example-image
      tags: latest
    environment:
      GOOGLE_APPLICATION_CREDENTIALS: /drone/src/.config/gcloud/application_default_credentials.json
      AWS_CREDENTIALS: /drone/src/.aws/credentials
      # OR
      AWS_ACCESS_KEY_ID: my-key-id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
```

Example `config.json` for `docker_cred_helpers_config`:

```json
{
  "auths": {
    "https://index.docker.io/v1/": {
      "auth": "username:password" // value must be base64 encoded together
    }
  },
  "credHelpers": {
    "gcr.io": "gcr",
    "us.gcr.io": "gcr",
    "aws_account_id.dkr.ecr.region.amazonaws.com": "ecr-login"
  }
}
```

[![Build Status](http://cloud.drone.io/api/badges/drone-plugins/drone-docker/status.svg)](http://cloud.drone.io/drone-plugins/drone-docker)
[![Gitter chat](https://badges.gitter.im/drone/drone.png)](https://gitter.im/drone/drone)
[![Join the discussion at https://discourse.drone.io](https://img.shields.io/badge/discourse-forum-orange.svg)](https://discourse.drone.io)
[![Drone questions at https://stackoverflow.com](https://img.shields.io/badge/drone-stackoverflow-orange.svg)](https://stackoverflow.com/questions/tagged/drone.io)
[![](https://images.microbadger.com/badges/image/plugins/docker.svg)](https://microbadger.com/images/plugins/docker "Get your own image badge on microbadger.com")
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-docker?status.svg)](http://godoc.org/github.com/drone-plugins/drone-docker)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-docker)](https://goreportcard.com/report/github.com/drone-plugins/drone-docker)

Drone plugin uses Docker-in-Docker to build and publish Docker images to a container registry. For the usage information and a listing of the available options please take a look at [the docs](http://plugins.drone.io/drone-plugins/drone-docker/).

## Build

Build the binaries with the following commands:

```console
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on

go build -v -a -tags netgo -o release/linux/amd64/drone-docker ./cmd/drone-docker
go build -v -a -tags netgo -o release/linux/amd64/drone-gcr ./cmd/drone-gcr
go build -v -a -tags netgo -o release/linux/amd64/drone-ecr ./cmd/drone-ecr
go build -v -a -tags netgo -o release/linux/amd64/drone-acr ./cmd/drone-acr
go build -v -a -tags netgo -o release/linux/amd64/drone-heroku ./cmd/drone-heroku
```

## Docker

Build the Docker images with the following commands:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/docker/Dockerfile.linux.amd64 --tag plugins/docker .

docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/gcr/Dockerfile.linux.amd64 --tag plugins/gcr .

docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/ecr/Dockerfile.linux.amd64 --tag plugins/ecr .

docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/acr/Dockerfile.linux.amd64 --tag plugins/acr .

docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/heroku/Dockerfile.linux.amd64 --tag plugins/heroku .
```

## Usage

> Notice: Be aware that the Docker plugin currently requires privileged capabilities, otherwise the integrated Docker daemon is not able to start.

```console
docker run --rm \
  -e PLUGIN_TAG=latest \
  -e PLUGIN_REPO=octocat/hello-world \
  -e DRONE_COMMIT_SHA=d8dbe4d94f15fe89232e0402c6e8a0ddf21af3ab \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  --privileged \
  plugins/docker --dry-run
```
