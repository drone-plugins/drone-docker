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

Publish an image with multiple tags:

```yaml
pipeline:
  docker:
    image: plugins/docker
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tags:
      - latest
      - 1.0.1
      - "1.0"
```

Build an image with additional arguments:

```yaml
pipeline:
  docker:
    image: plugins/docker
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    build_args:
      - HTTP_PROXY=http://yourproxy.com
```

## Using a custom registry

Please note that when using a custom registry (other than DockerHub) you will need to provide the registry URL and you will need to use a fully qualified repository name. For example:

```yaml
pipeline:
  docker:
    image: plugins/docker
    registry: http://registry.company.com
    repo: registry.company.com/my/image
```

## Caching

The Drone build environment is, by default, ephemeral meaning that you layers
are not saved between builds. There are two methods for caching your layers.

### Graph directory caching

This is the preferred method when using the `overlay` or `aufs` storage
drivers. Just use Drone's caching feature to backup and restore the directory
`/drone/docker`, as shown in the following example:

```yaml
pipeline:
  sftp_cache:
    image: plugins/sftp-cache
    restore: true
    mount: /drone/docker

  docker:
    image: plugins/docker
    storage_path: /drone/docker
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tags:
      - latest
      - "1.0.1"

  sftp_cache:
    image: plugins/sftp-cache
    rebuild: true
    mount: /drone/docker
```

## Troubleshooting

For detailed output you can set the `DOCKER_LAUNCH_DEBUG` environment variable
in your plugin configuration. This starts Docker with verbose logging enabled.

```yaml
pipeline:
  docker:
    environment:
      - DOCKER_LAUNCH_DEBUG=true
```

## Known Issues

There are known issues when attempting to run this plugin on CentOS, RedHat,
and Linux installations that do not have a supported storage driver installed.
You can check by running `docker info | grep 'Storage Driver:'` on your host
machine. If the storage driver is not `aufs` or `overlay` you will need to
re-configure your host machine.

This error occurs when trying to use the default `aufs` storage Driver but aufs
is not installed:

```
level=fatal msg="Error starting daemon: error initializing graphdriver: driver not supported
```

This error occurs when trying to use the `overlay` storage Driver but overlay
is not installed:

```
level=error msg="'overlay' not found as a supported filesystem on this host.
Please ensure kernel is new enough and has overlay support loaded."
level=fatal msg="Error starting daemon: error initializing graphdriver: driver not supported"
```

This error occurs when using CentOS or RedHat which default to the
`devicemapper` storage driver:

```
level=error msg="There are no more loopback devices available."
level=fatal msg="Error starting daemon: error initializing graphdriver: loopback mounting failed"
Cannot connect to the Docker daemon. Is 'docker -d' running on this host?
```

The above issue can be resolved by setting `storage_driver: vfs` in the
`.drone.yml` file. This may work, but will have very poor performance as
discussed [here](https://github.com/rancher/docker-from-scratch/issues/20).

This error occurs when using Debian wheezy or jessie and cgroups memory
features are not configured at the kernel level:

```
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/blkio cgroup blkio"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/perf_event cgroup perf_event"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/cpuset cgroup cpuset"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/cpu,cpuacct cgroup cpu,cpuacct"
time="2015-12-17T08:06:57Z" level=debug msg="Creating /sys/fs/cgroup/memory"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/memory cgroup memory"
time="2015-12-17T08:06:57Z" level=fatal msg="no such file or directory"
```

The above issue can be resolved by editing your `grub.cfg` and adding
`cgroup_enable=memory swapaccount=1` to you kernel image. This change should
look like that afterwards:

```
menuentry 'Debian GNU/Linux, avec Linux 3.16.0-0.bpo.4-amd64' --class debian --class gnu-linux --class gnu --class os {
        load_video
        insmod gzio
        insmod raid
        insmod mdraid09
        insmod part_msdos
        insmod part_msdos
        insmod part_msdos
        insmod ext2
        set root='(mduuid/dab6cffad124a3d7a4d2adc226fd5302)'
        search --no-floppy --fs-uuid --set=root a4085974-c507-4993-a9ed-bdc17e375cad
        echo    'Chargement de Linux 3.16.0-0.bpo.4-amd64 ...'
        linux   /boot/vmlinuz-3.16.0-0.bpo.4-amd64 root=/dev/md1 ro  cgroup_enable=memory swapaccount=1 quiet
        echo    'Chargement du disque mémoire initial ...'
        initrd  /boot/initrd.img-3.16.0-0.bpo.4-amd64
```
