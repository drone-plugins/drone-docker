FROM plugins/docker:linux-amd64

ADD release/linux/amd64/drone-acr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-acr"]
