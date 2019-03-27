FROM arm64v8/docker:18.09-dind

ADD release/linux/arm64/drone-docker /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker"]
