FROM plugins/docker:latest

RUN apk update && apk add ca-certificates

ADD release/linux/amd64/drone-docker-ecr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker-ecr"]
