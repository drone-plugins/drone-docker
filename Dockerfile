# Docker image for the docker plugin
#
#     docker build --rm=true -t plugins/docker .

FROM docker:1.11-dind

ADD drone-docker /bin/

ENTRYPOINT ["/bin/drone-docker"]
