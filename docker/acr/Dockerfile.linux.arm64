FROM plugins/docker:linux-arm64

ADD release/linux/arm64/drone-acr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-acr"]
