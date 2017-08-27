# docker build --rm -f docker/gcr/Dockerfile -t plugins/gcr .

FROM plugins/docker:latest

ADD release/linux/amd64/drone-docker-gcr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker-gcr"]
