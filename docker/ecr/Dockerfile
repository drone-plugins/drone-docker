FROM plugins/docker:latest

ADD release/linux/amd64/drone-ecr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-ecr"]
