Use the Docker plugin to build and push Docker images to a registry.
The following parameters are used to configure this plugin:

* `registry` - authenticates to this registry
* `username` - authenticates with this username
* `password` - authenticates with this password
* `email` - authenticates with this email
* `repo` - repository name for the image
* `tag` - repository tag for the image
* `file` - dockerfile to be used, defaults to Dockerfile
* `auth` - auth token for the registry
* `context` - the context path to use, defaults to root of the git repo
* `force_tag` - replace existing matched image tags
* `insecure` - enable insecure communication to this registry
* `mirror` - use a mirror registry instead of pulling images directly from the central Hub
* `bip` - use for pass bridge ip
* `dns` - set custom dns servers for the container
* `storage_driver` - use `aufs`, `devicemapper`, `btrfs` or `overlay` driver
* `save` - save image layers to the specified tar file (see [docker save](https://docs.docker.com/engine/reference/commandline/save/))
    * `destination` - absolute / relative destination path
    * `tag` - cherry-pick tags to save (optional)
* `load` - restore image layers from the specified tar file
* `build_args` - [build arguments](https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables-build-arg) to pass to `docker build`
* `mtu` - set custom mtu settings for the docker daemon

The following is a sample Docker configuration in your .drone.yml file:

```yaml
publish:
  docker:
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag: latest
    file: Dockerfile
    insecure: false
```

You may want to dynamically tag your image. Use the `$$BRANCH`, `$$COMMIT` and `$$BUILD_NUMBER` variables to tag your image with the branch, commit sha or build number:

```yaml
publish:
  docker:
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag: $$BRANCH
```

Or you may prefer to build an image with multiple tags:

```yaml
publish:
  docker:
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag:
      - latest
      - "1.0.1"
      - "1.0"
```

Note that in the above example we quote the version numbers. If the yaml parser interprets the value as a number it will cause a parsing error.

It's also possible to pass build arguments to docker:

```yaml
publish:
  docker:
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    build_args:
      - HTTP_PROXY=http://yourproxy.com
```

## Caching

The Drone build environment is, by default, ephemeral meaning that you layers are not saved between builds. There are two methods for caching your layers.

### Graph directory caching

This is the preferred method when using the `overlay` or `aufs` storage drivers. Just use Drone's caching feature to backup and restore the directory `/drone/docker`, as shown in the following example:

```yaml
publish:
  docker:
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag:
      - latest
      - "1.0.1"

cache:
  mount:
    - /drone/docker
```

NOTE: This probably won't work correctly with the `btrfs` driver, and it will be very inefficient with the `devicemapper` driver. Please make sure to use the `overlay` or `aufs` storage driver with this method.

### Layer Caching

The below example combines Drone's caching feature and Docker's `save` and `load` capabilities to cache and restore image layers between builds:

```yaml
publish:
  docker:
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag:
      - latest
      - "1.0.1"
    load: docker/image.tar
    save:
      destination: docker/image.tar
      tag: latest

cache:
  mount:
    - docker/image.tar
```

You might also want to create a `.dockerignore` file in your repo to exclude `image.tar` from Docker build context:

```
docker/*
```

In some cases caching will greatly improve build performance, however, the tradeoff is that caching Docker image layers may consume very large amounts of disk space.

## Troubleshooting

For detailed output you can set the `DOCKER_LAUNCH_DEBUG` environment variable in your plugin configuration. This starts Docker with verbose logging enabled.

```yaml
publish:
  docker:
    environment:
      - DOCKER_LAUNCH_DEBUG=true
```

## Known Issues

There are known issues when attempting to run this plugin on CentOS, RedHat, and Linux installations that do not have a supported storage driver installed. You can check by running `docker info | grep 'Storage Driver:'` on your host machine. If the storage driver is not `aufs` or `overlay` you will need to re-configure your host machine.

This error occurs when trying to use the default `aufs` storage Driver but aufs is not installed:

```
level=fatal msg="Error starting daemon: error initializing graphdriver: driver not supported
```

This error occurs when trying to use the `overlay` storage Driver but overlay is not installed:

```
level=error msg="'overlay' not found as a supported filesystem on this host.
Please ensure kernel is new enough and has overlay support loaded."
level=fatal msg="Error starting daemon: error initializing graphdriver: driver not supported"
```

This error occurs when using CentOS or RedHat which default to the `devicemapper` storage driver:

```
level=error msg="There are no more loopback devices available."
level=fatal msg="Error starting daemon: error initializing graphdriver: loopback mounting failed"
Cannot connect to the Docker daemon. Is 'docker -d' running on this host?
```

The above issue can be resolved by setting `storage_driver: vfs` in the `.drone.yml` file. This may work, but will have very poor performance as discussed [here](https://github.com/rancher/docker-from-scratch/issues/20).
