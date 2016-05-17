# drone-docker

Drone plugin can be used to build and publish Docker images to a container registry. For the usage information and a listing of the available options please take a look at [the docs](DOCS.md).

## Build

Build the binary with the following commands:

```
export GO15VENDOREXPERIMENT=1
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

go build -a -tags netgo
```

## Docker

Build the docker image with the following commands:

```
docker build --rm=true -t plugins/docker .
```

Please note incorrectly building the image for the correct x64 linux and with GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/go/bin/drone-docker' not found or does not exist..
```

## Usage

Build and publish from your current working directory:

```
docker run --rm \
  -e PLUGIN_TAG=latest \
  -e PLUGIN_REPO=octocat/hello-world \
  -e DRONE_COMMIT_SHA=d8dbe4d94f15fe89232e0402c6e8a0ddf21af3ab \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  --privileged \
  plugins/docker --dry-run
```
