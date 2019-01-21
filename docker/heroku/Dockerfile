FROM plugins/docker:latest

ADD release/linux/amd64/drone-heroku /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-heroku"]
