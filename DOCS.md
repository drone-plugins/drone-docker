Use the Docker plugin to build and push Docker images to a registry.
The following parameters are used to configure this plugin:

* **registry** - authenticates to this registry
* **username** - authenticates with this username
* **password** - authenticates with this password
* **email** - authenticates with this email
* **repo** - repository name for the image
* **tag** - repository tag for the image
* **insecure** - enable insecure communication to this registry
* **storage_driver** - use `aufs`, `devicemapper`, `btrfs` or `overlay` driver

The following is a sample Docker configuration in your .drone.yml file:

```yaml
pubish:
  docker:
    username: kevinbacon
    password: $$DOCKER_PASSWORD
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag: latest
    file: Dockerfile
    insecure: false
```

You may want to dynamically tag your image. Use the `$$BRANCH`, `$$COMMIT` and `$$BUILD_NUMBER` variables to tag your image with the branch, commit sha or build number:

```yaml
pubish:
  docker:
    username: kevinbacon
    password: $$DOCKER_PASSWORD
    email: kevin.bacon@mail.com
    repo: foo/bar
    tag: $$BRANCH
```
