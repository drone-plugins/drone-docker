FROM plugins/docker:linux-arm

ADD release/linux/arm/drone-gcr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-gcr"]
