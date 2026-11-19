FROM s390x/docker:dind

ENV DOCKER_HOST=unix:///var/run/docker.sock
ADD release/linux/s390x/drone-docker /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker"]
