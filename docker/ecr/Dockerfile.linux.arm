FROM plugins/docker:linux-arm

ADD release/linux/arm/drone-ecr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-ecr"]
