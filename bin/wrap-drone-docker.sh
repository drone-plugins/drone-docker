#!/bin/sh
set -eu

# attempt to detect JSON v. Baste64 encoding
if (echo "${TOKEN}" | jq -e '.project_id'); then
    DOCKER_PASSWORD="${TOKEN}"
else
    DOCKER_PASSWORD="$(echo "${TOKEN}" | base64 -d)"
fi

DOCKER_REGISTRY="${PLUGIN_REGISTRY:-gcr.io}"
DOCKER_USERNAME="_json_key"

# set variables for using Docker with GCR
export DOCKER_REGISTRY DOCKER_USERNAME DOCKER_PASSWORD

# invoke the docker plugin
exec /bin/drone-docker "$@"
