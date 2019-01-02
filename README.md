# drone-docker

Drone plugin to build and publish Docker images to a container registry.

### Special privilieges

This docker image is specified in Drone to run with Privilieged mode. If you fork this repository to run it on your own, you need to specify your docker image in the `DRONE_ESCALATE`, as the container needs to run in privilieged mode, otherwise it will not work! https://github.com/drone/drone/blob/master/cmd/drone-server/server.go

## Build

Build the binary with the following commands:

```
sh .drone.sh
```

## Docker

Build the Docker image with the following commands:

```
docker build --rm=true -f docker/Dockerfile -t plugins/docker .
```

## Usage

Execute from the working directory:

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
