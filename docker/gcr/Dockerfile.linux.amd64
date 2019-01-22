FROM plugins/docker:linux-amd64

ADD release/linux/amd64/drone-gcr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-gcr"]
