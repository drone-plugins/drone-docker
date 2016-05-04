Use the Docker plugin to build and push Docker images to a registry. The following parameters are used to configure this plugin:

* `registry` - authenticates to this registry
* `username` - authenticates with this username
* `password` - authenticates with this password
* `email` - authenticates with this email
* `repo` - repository name for the image
* `tag` - repository tag for the image
* `file` - dockerfile to be used, defaults to Dockerfile
* `context` - the context path to use, defaults to root of the git repo
* `insecure` - enable insecure communication to this registry
* `mirror` - use a mirror registry instead of pulling images directly from the central Hub
* `bip` - use for pass bridge ip
* `dns` - set custom dns servers for the container
* `storage_driver` - use `aufs`, `devicemapper`, `btrfs` or `overlay` driver
* `storage_path` - location of docker daemon storage on disk
* `build_args` - [build arguments](https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables-build-arg) to pass to `docker build`

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

Publish and image with multiple tags:

```yaml
publish:
  docker:
    username: kevinbacon
    password: pa55word
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag:
      - latest
      - 1.0.1
      - "1.0"
```

Build an image with arguments:

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
    storage_path: /drone/docker
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

This error occurs when using Debian wheezy or jessie and cgroups memory features are not configured at the kernel level:

```
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/blkio cgroup blkio"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/perf_event cgroup perf_event"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/cpuset cgroup cpuset"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/cpu,cpuacct cgroup cpu,cpuacct"
time="2015-12-17T08:06:57Z" level=debug msg="Creating /sys/fs/cgroup/memory"
time="2015-12-17T08:06:57Z" level=debug msg="Mounting none /sys/fs/cgroup/memory cgroup memory"
time="2015-12-17T08:06:57Z" level=fatal msg="no such file or directory"
```

The above issue can be resolved by editing your `grub.cfg` and adding `cgroup_enable=memory swapaccount=1` to you kernel image. This change should look like that afterwards:

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
