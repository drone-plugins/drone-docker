FROM plugins/docker:latest

ADD release/linux/amd64/drone-docker-ecr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker-ecr"]
