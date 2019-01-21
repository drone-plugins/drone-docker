FROM plugins/docker:linux-arm

ADD release/linux/arm/drone-heroku /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-heroku"]
