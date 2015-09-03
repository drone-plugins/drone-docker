# drone-docker
Drone plugin for publishing Docker images


## Docker

Build the Docker container:

```sh
docker build --rm=true -t plugins/drone-docker .
```

Build and Publish a Docker container

```sh
docker run -i --privileged -v $(pwd):/drone/src plugins/drone-docker <<EOF
{
	"workspace": {
		"path": "/drone/src"
	},
	"build" : {
		"number": 1,
		"head_commit": {
			"sha": "9f2849d5",
			"branch": "master",
			"ref": "refs/heads/master"
		}
	},
	"vargs": {
		"username": "kevinbacon",
		"password": "pa$$word", 
		"email": "foo@bar.com", 
		"repo": "foo/bar",
		"storage_driver": "aufs"
	}
}
EOF
```
