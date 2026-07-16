# drone-docker

[![Build Status](http://cloud.drone.io/api/badges/drone-plugins/drone-docker/status.svg)](http://cloud.drone.io/drone-plugins/drone-docker)
[![Gitter chat](https://badges.gitter.im/drone/drone.png)](https://gitter.im/drone/drone)
[![Join the discussion at https://discourse.drone.io](https://img.shields.io/badge/discourse-forum-orange.svg)](https://discourse.drone.io)
[![Drone questions at https://stackoverflow.com](https://img.shields.io/badge/drone-stackoverflow-orange.svg)](https://stackoverflow.com/questions/tagged/drone.io)
[![](https://images.microbadger.com/badges/image/plugins/docker.svg)](https://microbadger.com/images/plugins/docker "Get your own image badge on microbadger.com")
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-docker?status.svg)](http://godoc.org/github.com/drone-plugins/drone-docker)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-docker)](https://goreportcard.com/report/github.com/drone-plugins/drone-docker)

Drone plugin uses Docker-in-Docker to build and publish Docker images to a container registry. For the usage information and a listing of the available options please take a look at [the docs](http://plugins.drone.io/drone-plugins/drone-docker/).

### Git Leaks

Run the following script to install git-leaks support to this repo.
```
chmod +x ./git-hooks/install.sh
./git-hooks/install.sh
```

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
go build -v -a -tags netgo -o release/linux/amd64/drone-gar ./cmd/drone-gar
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
  
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/gar/Dockerfile.linux.amd64 --tag plugins/gar .
```

## Usage

> Notice: Be aware that the Docker plugin currently requires privileged capabilities, otherwise the integrated Docker daemon is not able to start.

### Using Docker buildkit Secrets

```yaml
kind: pipeline
name: default

steps:
- name: build dummy docker file and publish
  image: plugins/docker
  pull: never
  settings:
    repo: tphoney/test
    tags: latest
    secret: id=mysecret,src=secret-file
    username:
      from_secret: docker_username
    password:
      from_secret: docker_password
```

Using a dockerfile that references the secret-file 

```bash
# syntax=docker/dockerfile:1.2

FROM alpine

# shows secret from default secret location:
RUN --mount=type=secret,id=mysecret cat /run/secrets/mysecret
```

and a secret file called secret-file

```
COOL BANANAS
```


### Trusting a custom / egress-proxy CA (`HARNESS_CA_PATH`)

When builds run behind a TLS-intercepting egress proxy (e.g. Harness egress
control), the proxy re-signs upstream TLS with its own CA. The Docker daemon and
its embedded BuildKit verify registry TLS against the **host system trust
store**, so base-image pulls fail with
`x509: certificate signed by unknown authority` unless that CA is trusted.

Set the `HARNESS_CA_PATH` environment variable to a PEM CA (bundle) file and the
plugin installs it into the host trust store **before any HTTPS** (registry
wrapper auth such as GCP STS, and before starting the Docker daemon):

```yaml
steps:
- name: build and push
  image: plugins/docker
  pull: never
  environment:
    HARNESS_CA_PATH: /etc/harness-certs/ca.crt
  settings:
    repo: octocat/hello-world
    tags: latest
```

Behavior:

- **Linux** — installs into the distro anchor dir
  (`/usr/local/share/ca-certificates` on Debian/Ubuntu/Alpine,
  `/etc/pki/ca-trust/source/anchors` on RHEL/CentOS/Fedora) and runs
  `update-ca-certificates` / `update-ca-trust`. It also appends the CA directly
  to the consolidated bundle (`/etc/ssl/certs/ca-certificates.crt` or the RHEL
  equivalent) so trust holds even on minimal images without a refresh tool.
- **Windows** — imports the CA into `LocalMachine\Root` via the Windows
  CryptoAPI (`crypt32.dll`). This works on Nano Server images that lack
  `certutil` / PowerShell.
- **macOS / other** — logged no-op (not currently supported).

Notes:

- Registry wrappers (`plugins/gcr`, `gar`, `acr`, `ecr`) call the same install
  before their pre-docker HTTPS auth (e.g. `sts.googleapis.com` token exchange).
- It is **best-effort and idempotent**: when `HARNESS_CA_PATH` is unset, the file
  is missing, or the file is empty, the plugin logs and continues. Re-running the
  step will not duplicate the CA in the bundle.
- The CA is trusted for the daemon's own registry TLS (base-image pulls, cache
  endpoints) and for Go TLS clients in the plugin process. To make the CA
  available to `RUN` steps inside the build, also pass it as needed via the
  Dockerfile / build args.
- This is separate from the proxy build args
  (`http_proxy`/`https_proxy`/`no_proxy`), which the plugin already injects from
  the environment (including `HARNESS_`-prefixed variants).

### Running from the CLI

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

### GAR (Google Artifact Registry)

```yaml
kind: pipeline
name: default
type: docker

steps:
  - name: push-to-gar
    image: plugins/gar
    pull: never
    settings:
      tag: latest
      repo: project-id/repo/image-name
      location: us
      json_key:
        from_secret: gcr_json_key
```

### GAR (Google Artifact Registry) using workload identity (OIDC)

```yaml
steps:
  - name: push-to-gar
    image: plugins/gar
    pull: never
    settings:
      tag: latest
      repo: project-id/repo/image-name
      location: europe
      project_number: project-number
      pool_id: workload identity pool id
      provider_id: workload identity provider id
      service_account_email: service account email
      oidc_token_id:
        from_secret: token 
```

## Developer Notes

- When updating the base image, you will need to update for each architecture and OS.
- Arm32 base images are no longer being updated.

## Release procedure

Run the changelog generator.

```BASH
docker run -it --rm -v "$(pwd)":/usr/local/src/your-app githubchangeloggenerator/github-changelog-generator -u drone-plugins -p drone-docker -t <secret github token>
```

You can generate a token by logging into your GitHub account and going to Settings -> Personal access tokens.

Next we tag the PR's with the fixes or enhancements labels. If the PR does not fufil the requirements, do not add a label.

Run the changelog generator again with the future version according to semver.

```BASH
docker run -it --rm -v "$(pwd)":/usr/local/src/your-app githubchangeloggenerator/github-changelog-generator -u drone-plugins -p drone-docker -t <secret token> --future-release v1.0.0
```

Create your pull request for the release. Get it merged then tag the release.

