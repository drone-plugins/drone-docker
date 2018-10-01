# docker build --rm -f docker/gcr/Dockerfile -t plugins/gcr .
FROM golang:1.11
WORKDIR /go/src/github.com/drone-plugins/drone-docker
ADD . .
RUN go get -u github.com/golang/dep/cmd/dep && \
    dep ensure && \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
        -tags netgo -o /bin/drone-gcr \
        github.com/drone-plugins/drone-docker/cmd/drone-docker-gcr

FROM plugins/docker:latest

COPY --from=0 /bin/drone-gcr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-gcr"]
