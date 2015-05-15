Use the Docker plugin to build and push Docker images to a registry.
The following parameters are used to configuration this plugin:

* **registry** - authenticates to this registry
* **username** - authenticates with this username
* **password** - authenticates with this password
* **email** - authenticates with this email
* **repo** - repository name for the image
* **tag** - repository tag for the image

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
```
