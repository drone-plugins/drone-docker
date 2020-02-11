FROM golang:1.13 AS builder
ENV GOARCH=amd64 GOOS=linux CGO_ENABLED=0 GO111MODULE=on
WORKDIR /build
COPY . .

RUN go build -tags netgo ./cmd/drone-ecr
RUN go build -tags netgo ./cmd/drone-docker

FROM docker:18.09.0-dind

RUN apk update && apk upgrade && \
    apk add --no-cache git

COPY --from=builder /build/drone-ecr /bin/
COPY --from=builder /build/drone-docker /bin/

ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-ecr"]