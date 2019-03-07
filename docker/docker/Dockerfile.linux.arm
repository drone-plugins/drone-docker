FROM arm32v6/docker:18.09-dind

ADD release/linux/arm/drone-docker /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker"]
