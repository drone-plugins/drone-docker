local PipelineTesting = {
  kind: "pipeline",
  name: "testing",
  platform: {
    os: "linux",
    arch: "amd64",
  },
  steps: [
    {
      name: "vet",
      image: "golang:1.11",
      pull: "always",
      environment: {
        GO111MODULE: "on",
      },
      commands: [
        "go vet ./...",
      ],
    },
    {
      name: "test",
      image: "golang:1.11",
      pull: "always",
      environment: {
        GO111MODULE: "on",
      },
      commands: [
        "go test -cover ./...",
      ],
    },
  ],
  trigger: {
    branch: [ "master" ],
  },
};

local PipelineBuild(os="linux", arch="amd64") = {
  kind: "pipeline",
  name: os + "-" + arch,
  platform: {
    os: os,
    arch: arch,
  },
  steps: [
    {
      name: "build-docker",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-docker ./cmd/drone-docker",
      ],
      when: {
        event: [ "push", "pull_request" ],
      },
    },
    {
      name: "build-docker",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.version=${DRONE_TAG##v} -X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-docker ./cmd/drone-docker",
      ],
      when: {
        event: [ "tag" ],
      },
    },
    {
      name: "dryrun-docker",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        dry_run: true,
        tags: os + "-" + arch,
        dockerfile: "docker/docker/Dockerfile." + os + "." + arch,
        repo: "plugins/docker",
        username: { "from_secret": "docker_username" },
        password: { "from_secret": "docker_password" },
      },
      when: {
        event: [ "pull_request" ],
      },
    },
    {
      name: "publish-docker",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        auto_tag: true,
        auto_tag_suffix: os + "-" + arch,
        dockerfile: "docker/docker/Dockerfile." + os + "." + arch,
        repo: "plugins/docker",
        username: { "from_secret": "docker_username" },
        password: { "from_secret": "docker_password" },
      },
      when: {
        event: [ "push", "tag" ],
      },
    },
    {
      name: "build-heroku",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-heroku ./cmd/drone-heroku",
      ],
      when: {
        event: [ "push", "pull_request" ],
      },
    },
    {
      name: "build-heroku",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.version=${DRONE_TAG##v} -X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-heroku ./cmd/drone-heroku",
      ],
      when: {
        event: [ "tag" ],
      },
    },
    {
      name: "dryrun-heroku",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        dry_run: true,
        tags: os + "-" + arch,
        herokufile: "docker/heroku/Dockerfile." + os + "." + arch,
        repo: "plugins/heroku",
        username: { "from_secret": "heroku_username" },
        password: { "from_secret": "heroku_password" },
      },
      when: {
        event: [ "pull_request" ],
      },
    },
    {
      name: "publish-heroku",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        auto_tag: true,
        auto_tag_suffix: os + "-" + arch,
        herokufile: "docker/heroku/Dockerfile." + os + "." + arch,
        repo: "plugins/heroku",
        username: { "from_secret": "heroku_username" },
        password: { "from_secret": "heroku_password" },
      },
      when: {
        event: [ "push", "tag" ],
      },
    },
    {
      name: "build-gcr",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-gcr ./cmd/drone-gcr",
      ],
      when: {
        event: [ "push", "pull_request" ],
      },
    },
    {
      name: "build-gcr",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.version=${DRONE_TAG##v} -X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-gcr ./cmd/drone-gcr",
      ],
      when: {
        event: [ "tag" ],
      },
    },
    {
      name: "dryrun-gcr",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        dry_run: true,
        tags: os + "-" + arch,
        gcrfile: "docker/gcr/Dockerfile." + os + "." + arch,
        repo: "plugins/gcr",
        username: { "from_secret": "gcr_username" },
        password: { "from_secret": "gcr_password" },
      },
      when: {
        event: [ "pull_request" ],
      },
    },
    {
      name: "publish-gcr",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        auto_tag: true,
        auto_tag_suffix: os + "-" + arch,
        gcrfile: "docker/gcr/Dockerfile." + os + "." + arch,
        repo: "plugins/gcr",
        username: { "from_secret": "gcr_username" },
        password: { "from_secret": "gcr_password" },
      },
      when: {
        event: [ "push", "tag" ],
      },
    },
    {
      name: "build-ecr",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-ecr ./cmd/drone-ecr",
      ],
      when: {
        event: [ "push", "pull_request" ],
      },
    },
    {
      name: "build-ecr",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.version=${DRONE_TAG##v} -X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-ecr ./cmd/drone-ecr",
      ],
      when: {
        event: [ "tag" ],
      },
    },
    {
      name: "dryrun-ecr",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        dry_run: true,
        tags: os + "-" + arch,
        dockerfile: "docker/ecr/Dockerfile." + os + "." + arch,
        repo: "plugins/ecr",
        username: { "from_secret": "ecr_username" },
        password: { "from_secret": "ecr_password" },
      },
      when: {
        event: [ "pull_request" ],
      },
    },
    {
      name: "publish-ecr",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        auto_tag: true,
        auto_tag_suffix: os + "-" + arch,
        dockerfile: "docker/ecr/Dockerfile." + os + "." + arch,
        repo: "plugins/ecr",
        username: { "from_secret": "ecr_username" },
        password: { "from_secret": "ecr_password" },
      },
      when: {
        event: [ "push", "tag" ],
      },
    },
  ],
  depends_on: [
    "testing",
  ],
  trigger: {
    branch: [ "master" ],
  },
};

local PipelineNotifications = {
  kind: "pipeline",
  name: "notifications",
  platform: {
    os: "linux",
    arch: "amd64",
  },
  steps: [
    {
      name: "manifest-docker",
      image: "plugins/manifest:1",
      pull: "always",
      settings: {
        username: { "from_secret": "docker_username" },
        password: { "from_secret": "docker_password" },
        spec: "docker/docker/manifest.tmpl",
        ignore_missing: true,
      },
    },
    {
      name: "microbadger-docker",
      image: "plugins/webhook:1",
      pull: "always",
      settings: {
        url: { "from_secret": "microbadger_docker" },
      },
    },
    {
      name: "manifest-heroku",
      image: "plugins/manifest:1",
      pull: "always",
      settings: {
        username: { "from_secret": "heroku_username" },
        password: { "from_secret": "heroku_password" },
        spec: "docker/heroku/manifest.tmpl",
        ignore_missing: true,
      },
    },
    {
      name: "microbadger-heroku",
      image: "plugins/webhook:1",
      pull: "always",
      settings: {
        url: { "from_secret": "microbadger_heroku" },
      },
    },
    {
      name: "manifest-gcr",
      image: "plugins/manifest:1",
      pull: "always",
      settings: {
        username: { "from_secret": "gcr_username" },
        password: { "from_secret": "gcr_password" },
        spec: "docker/gcr/manifest.tmpl",
        ignore_missing: true,
      },
    },
    {
      name: "microbadger-gcr",
      image: "plugins/webhook:1",
      pull: "always",
      settings: {
        url: { "from_secret": "microbadger_gcr" },
      },
    },
    {
      name: "manifest-ecr",
      image: "plugins/manifest:1",
      pull: "always",
      settings: {
        username: { "from_secret": "ecr_username" },
        password: { "from_secret": "ecr_password" },
        spec: "docker/ecr/manifest.tmpl",
        ignore_missing: true,
      },
    },
    {
      name: "microbadger-ecr",
      image: "plugins/webhook:1",
      pull: "always",
      settings: {
        url: { "from_secret": "microbadger_ecr" },
      },
    },
  ],
  depends_on: [
    "linux-amd64",
    "linux-arm64",
    "linux-arm",
  ],
  trigger: {
    branch: [ "master" ],
    event: [ "push", "tag" ],
  },
};

[
  PipelineTesting,
  PipelineBuild("linux", "amd64"),
  PipelineBuild("linux", "arm64"),
  PipelineBuild("linux", "arm"),
  PipelineNotifications,
]
