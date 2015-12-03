# Docker image for the docker plugin
#
#     docker build --rm=true -t plugins/drone-docker .

FROM rancher/docker:1.9.1

ADD drone-docker /go/bin/
VOLUME /var/lib/docker
ENTRYPOINT ["/usr/bin/dockerlaunch", "/go/bin/drone-docker"]
