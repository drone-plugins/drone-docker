# Docker image for the docker plugin
#
#     docker build --rm=true -t plugins/docker .

FROM rancher/docker:v1.10.2

ADD drone-docker /go/bin/
VOLUME /var/lib/docker
ENTRYPOINT ["/usr/bin/dockerlaunch", "/go/bin/drone-docker"]
