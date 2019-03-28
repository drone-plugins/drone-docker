FROM plugins/docker:linux-arm

ADD release/linux/arm/drone-acr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-acr"]
