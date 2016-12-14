---
date: 2016-01-01T00:00:00+00:00
title: Docker
author: drone-plugins
tags: [ publish, docker ]
repo: drone-plugins/drone-docker
image: plugins/drone
---

The Docker plugin can be used to build and publish images to the Docker registry. The following pipeline configuration uses the Docker plugin to build and publish Docker images:

```yaml
pipeline:
  docker:
    image: plugins/docker
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tags: latest
```

Example configuration using multiple tags:

```diff
pipeline:
  docker:
    image: plugins/docker
    repo: foo/bar
-   tags: latest
+   tags:
+     - latest
+     - 1.0.1
+     - 1.0
```

Example configuration using build arguments:

```diff
publish:
  docker:
    image: plugins/docker
    repo: foo/bar
+   build_args:
+     - HTTP_PROXY=http://yourproxy.com
```

Example configuration using alternate Dockerfile:

```diff
publish:
  docker:
    image: plugins/docker
    repo: foo/bar
-   dockerfile: Dockerfile
+   dockerfile: path/to/Dockerfile
```

Example configuration using a custom registry:

```diff
pipeline:
  docker:
    image: plugins/docker
-   repo: foo/bar
+   repo: index.company.com/foo/bar
+   registry: index.company.com
```

Example configuration using inline credentials:

```diff
pipeline:
  docker:
    image: plugins/docker
+   username: kevinbacon
+   password: pa55word
    repo: foo/bar
```

# Secrets

The Docker plugin supports reading credentials from the Drone secret store. This is strongly recommended instead of storing credentials in the pipeline configuration in plain text.

```diff
pipeline:
  docker:
    image: plugins/docker
-   username: kevinbacon
-   password: pa55word
    repo: foo/bar
```

Use the command line utility to add secrets to the store:

```nohighlight
drone secret add --image=plugins/docker \
    octocat/hello-world DOCKER_USERNAME kevinbacon

drone secret add --image=plugins/docker \
    octocat/hello-world DOCKER_PASSWORD pa55word
```

Don't forget to sign the Yaml after making changes:

```nohighlight
drone sign octocat/hello-world
```

# Secret Reference

DOCKER_USERNAME
: docker registry username

DOCKER_PASSWORD
: docker registry password

DOCKER_EMAIL
: docker registry email

# Parameter Reference

registry
: authenticates to this registry

username
: authenticates with this username

password
: authenticates with this password

email
: authenticates with this email

repo
: repository name for the image

tags
: repository tag for the image

dockerfile
: dockerfile to be used, defaults to Dockerfile

auth
: auth token for the registry

context
: the context path to use, defaults to root of the git repo

force_tag=false
: replace existing matched image tags

insecure=false
: enable insecure communication to this registry

mirror
: use a mirror registry instead of pulling images directly from the central Hub

bip=false
: use for pass bridge ip

dns
: set custom dns servers for the container

storage_driver
: supports `aufs`, `overlay` or `vfs` drivers

build_args
: custom arguments passed to docker build
