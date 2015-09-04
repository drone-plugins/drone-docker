# Docker image for the docker plugin
#
#     docker build --rm=true -t plugins/drone-docker .

FROM rancher/docker:1.8.1

ADD drone-docker /go/bin/
VOLUME /var/lib/docker
ENTRYPOINT ["/go/bin/drone-docker"]
