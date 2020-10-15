golang_image = "golang:1.15"

def main(ctx):
    before = testing(ctx)

    stages = [
        linux(ctx, "amd64"),
        linux(ctx, "arm64"),
        linux(ctx, "arm"),
    ]

    after = manifest(ctx) + gitter(ctx)

    for b in before:
        for s in stages:
            s["depends_on"].append(b["name"])

    for s in stages:
        for a in after:
            a["depends_on"].append(s["name"])

    return before + stages + after

def testing(ctx):
    step_volumes = [
        {
            "name": "gopath",
            "path": "/go",
        },
    ]

    return [
        {
            "kind": "pipeline",
            "type": "docker",
            "name": "testing",
            "platform": {
                "os": "linux",
                "arch": "amd64",
            },
            "steps": [
                {
                    "name": "staticcheck",
                    "image": golang_image,
                    "pull": "always",
                    "commands": [
                        "go run honnef.co/go/tools/cmd/staticcheck ./...",
                    ],
                    "volumes": step_volumes,
                },
                {
                    "name": "lint",
                    "image": golang_image,
                    "pull": "always",
                    "commands": [
                        "go run golang.org/x/lint/golint -set_exit_status ./...",
                    ],
                    "volumes": step_volumes,
                },
                {
                    "name": "vet",
                    "image": golang_image,
                    "commands": [
                        "go vet ./...",
                    ],
                    "volumes": step_volumes,
                },
                {
                    "name": "test",
                    "image": golang_image,
                    "commands": [
                        "go test -cover ./...",
                    ],
                    "volumes": step_volumes,
                },
            ],
            "volumes": [
                {
                    "name": "gopath",
                    "temp": {},
                },
            ],
            "trigger": {
                "ref": [
                    "refs/heads/master",
                    "refs/tags/**",
                    "refs/pull/**",
                ],
            },
        },
    ]

def linux(ctx, arch):
    steps = [
        {
            "name": "environment",
            "image": golang_image,
            "pull": "always",
            "environment": {
                "CGO_ENABLED": "0",
            },
            "commands": [
                "go version",
                "go env",
            ],
        },
    ]

    steps.extend(linux_build(ctx, arch, "docker"))
    steps.extend(linux_build(ctx, arch, "acr"))
    steps.extend(linux_build(ctx, arch, "ecr"))
    steps.extend(linux_build(ctx, arch, "gcr"))
    steps.extend(linux_build(ctx, arch, "heroku"))

    return {
        "kind": "pipeline",
        "type": "docker",
        "name": "linux-%s" % (arch),
        "platform": {
            "os": "linux",
            "arch": arch,
        },
        "steps": steps,
        "depends_on": [],
        "trigger": {
            "ref": [
                "refs/heads/master",
                "refs/tags/**",
                "refs/pull/**",
            ],
        },
    }

def linux_build(ctx, arch, name):
    docker = {
        "dockerfile": "docker/%s/Dockerfile.linux.%s" % (name, arch),
        "repo": "plugins/%s" % (name),
        "username": {
            "from_secret": "docker_username",
        },
        "password": {
            "from_secret": "docker_password",
        },
    }

    if ctx.build.event == "pull_request":
        docker.update({
            "dry_run": True,
            "tags": "linux-%s" % (arch),
        })
    else:
        docker.update({
            "auto_tag": True,
            "auto_tag_suffix": "linux-%s" % (arch),
        })

    if ctx.build.event == "tag":
        build = [
            'go build -v -ldflags "-X main.version=%s" -a -tags netgo -o release/linux/%s/drone-npm ./cmd/drone-npm' % (ctx.build.ref.replace("refs/tags/v", ""), arch),
        ]
    else:
        build = [
            'go build -v -ldflags "-X main.version=%s" -a -tags netgo -o release/linux/%s/drone-npm ./cmd/drone-npm' % (ctx.build.commit[0:8], arch),
        ]

    return [
        {
            "name": "build-%s" % (name),
            "image": golang_image,
            "environment": {
                "CGO_ENABLED": "0",
            },
            "commands": build,
        },
        {
            "name": "executable-%s" % (name),
            "image": golang_image,
            "commands": [
                "./release/linux/%s/drone-%s --help" % (arch, name),
            ],
        },
        {
            "name": "docker-%s" % (name),
            "image": "plugins/docker",
            "pull": "always",
            "settings": docker,
        },
    ]

def manifest(ctx):
    steps = []

    steps.extend(manifest_build(ctx, "docker"))
    steps.extend(manifest_build(ctx, "acr"))
    steps.extend(manifest_build(ctx, "ecr"))
    steps.extend(manifest_build(ctx, "gcr"))
    steps.extend(manifest_build(ctx, "heroku"))

    return [
        {
            "kind": "pipeline",
            "type": "docker",
            "name": "manifest",
            "steps": steps,
            "depends_on": [],
            "trigger": {
                "ref": [
                    "refs/heads/master",
                    "refs/tags/**",
                ],
            },
        },
    ]

def manifest_build(ctx, name):
    return [
        {
            "name": "manifest-%s" % (name),
            "image": "plugins/manifest",
            "pull": "always",
            "settings": {
                "auto_tag": "true",
                "username": {
                    "from_secret": "docker_username",
                },
                "password": {
                    "from_secret": "docker_password",
                },
                "spec": "docker/%s/manifest.tmpl" % (name),
                "ignore_missing": "true",
            },
        },
        {
            "name": "microbadger-%s" % (name),
            "image": "plugins/webhook",
            "pull": "always",
            "settings": {
                "urls": {
                    "from_secret": "microbadger_url",
                },
            },
        },
    ]

def gitter(ctx):
    return [
        {
            "kind": "pipeline",
            "type": "docker",
            "name": "gitter",
            "clone": {
                "disable": True,
            },
            "steps": [
                {
                    "name": "gitter",
                    "image": "plugins/gitter",
                    "pull": "always",
                    "settings": {
                        "webhook": {
                            "from_secret": "gitter_webhook",
                        },
                    },
                },
            ],
            "depends_on": [
                "manifest",
            ],
            "trigger": {
                "ref": [
                    "refs/heads/master",
                    "refs/tags/**",
                ],
                "status": [
                    "failure",
                ],
            },
        },
    ]
