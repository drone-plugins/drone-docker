# docker build --rm -f docker/heroku/Dockerfile -t plugins/heroku .

FROM plugins/docker:latest

ADD release/linux/amd64/drone-docker-heroku /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker-heroku"]
