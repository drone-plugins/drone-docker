# Docker image for the docker plugin
#
#     docker build --rm=true -t plugins/drone-docker .

FROM ubuntu:14.04

RUN apt-get update -qq                               \
	&& apt-get -y install curl                       \
	&& curl -sSL https://get.docker.com/ubuntu/ | sh \
	&& rm -rf /var/lib/apt/lists/*

ADD drone-docker /go/bin/

ENTRYPOINT ["/go/bin/drone-docker"]
