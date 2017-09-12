#! /bin/bash
##
# Created for lflare/drone-hyper
# By Amos (LFlare) Ng <amosng1@gmail.com
##
# Run original entrypoint
/entrypoint.sh

# Relink to original locations
ln -sf /bin/docker /usr/local/bin/docker
ln -sf /bin/dockerd /usr/local/bin/dockerd

# Run docker daemon
/usr/bin/dockerd 2>&1 > /dev/null &

# Run commands
exec "$@"
