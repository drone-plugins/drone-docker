drone-plugin-go
===============

This is a package with simple support for writing Drone plugins in Go.

## Overview

Plugins are executable files run by Drone to customize the build lifecycle. Plugins receive input data from `stdin` (or `arg[1]`) and write the results to `stdout`

```sh
./slack-plugin <<EOF
{
    "repo" : {
        "owner": "foo",
        "name": "bar",
        "full_name": "foo/bar"
    },
    "build" : {
        "number": 1
        "status": "success",
        "started_at": 1421029603,
        "finished_at": 1421029813,
        "head_commit" : {
            "sha": "9f2849d5",
            "ref": "refs/heads/master"
            "branch": "master",
            "message": "Update the Readme",
            "author": {
                "login": "johnsmith"
                "email": "john.smith@gmail.com",
            }
        }
    },
    "job" : {
        "number": 1,
        "status": "success",
        "started_at": 1421029603,
        "finished_at": 1421029813,
        "exit_code": 0,
        "environment": { "GO_VERSION": "1.4" }
    }
    "clone" : {
        "branch": "master",
        "remote": "git://github.com/drone/drone",
        "dir": "/drone/src/github.com/drone/drone",
        "ref": "refs/heads/master",
        "sha": "436b7a6e2abaddfd35740527353e78a227ddcb2c"
    },
    "vargs": {
        "webhook_url": "https://hooks.slack.com/services/...",
        "username": "drone",
        "channel": "#dev"
    }
}
EOF
```

Use this `plugin` package to retrieve and parse input parameters:

```Go
var repo = plugin.Repo{}
var build = plugin.Build{}
var slack = struct {
    URL      string `json:"webhook_url"`
    Username string `json:"username"`
    Channel  string `json:"channel"`
}{}

plugin.Param("repo", &repo)
plugin.Param("build", &build)
plugin.Param("vargs", &slack)
plugin.Parse()
```

Note that your plugin configuration data (declared in the `.drone.yml` file) will be provided in the `vargs` section of the JSON input.

### Shared Volumes

The repository clone directory (specified in the `clone.dir` input parameter) will be shared across all plugins as a [container volume](https://docs.docker.com/userguide/dockervolumes/#creating-and-mounting-a-data-volume-container). This means that any files in your repository directory or subdirectories are accessible to plugins. This is useful for plugins that analyze or archive files, such as an S3 plugin.

### Publishing

Drone plugins are distributed as Docker images. We therefore recommend publishing your plugins to [Docker Hub](https://index.docker.io).

The `ENTRYPOINT` must be defined and must point to your executable file. The `CMD` section will be overridden by Drone and will be used to send the JSON encoded data in `arg[1]`. An example Dockerfile for your plugin might look like this:

```Dockerfile
# Docker image for Drone's git-clone plugin
#
#     docker build -t drone/drone-clone-git .

FROM library/golang:1.4

# copy the local src files to the container's workspace.
ADD . /go/src/github.com/drone/drone-clone-git/

# build the git-clone plugin inside the container.
RUN go get github.com/drone/drone-clone-git/... && \
    go install github.com/drone/drone-clone-git

# run the git-clone plugin when the container starts
ENTRYPOINT ["/go/bin/drone-clone-git"]
```
