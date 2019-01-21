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

local PipelineBuild(binary="docker", os="linux", arch="amd64") = {
  kind: "pipeline",
  name: os + "-" + arch + "-" + binary,
  platform: {
    os: os,
    arch: arch,
  },
  steps: [
    {
      name: "build-push",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-" + binary + " ./cmd/drone-" + binary,
      ],
      when: {
        event: [ "push", "pull_request" ],
      },
    },
    {
      name: "build-tag",
      image: "golang:1.11",
      pull: "always",
      environment: {
        CGO_ENABLED: "0",
        GO111MODULE: "on",
      },
      commands: [
        "go build -v -ldflags \"-X main.version=${DRONE_TAG##v} -X main.build=${DRONE_BUILD_NUMBER}\" -a -tags netgo -o release/" + os + "/" + arch + "/drone-" + binary + " ./cmd/drone-" + binary,
      ],
      when: {
        event: [ "tag" ],
      },
    },
    {
      name: "dryrun",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        dry_run: true,
        tags: os + "-" + arch,
        dockerfile: "docker/" + binary + "/Dockerfile." + os + "." + arch,
        repo: "plugins/" + binary,
        username: { "from_secret": "docker_username" },
        password: { "from_secret": "docker_password" },
      },
      when: {
        event: [ "pull_request" ],
      },
    },
    {
      name: "publish",
      image: "plugins/docker:" + os + "-" + arch,
      pull: "always",
      settings: {
        auto_tag: true,
        auto_tag_suffix: os + "-" + arch,
        dockerfile: "docker/" + binary + "/Dockerfile." + os + "." + arch,
        repo: "plugins/" + binary,
        username: { "from_secret": "docker_username" },
        password: { "from_secret": "docker_password" },
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

local PipelineNotifications(binary="docker") = {
  kind: "pipeline",
  name: "notifications-" + binary,
  platform: {
    os: "linux",
    arch: "amd64",
  },
  steps: [
    {
      name: "manifest",
      image: "plugins/manifest:1",
      pull: "always",
      settings: {
        username: { "from_secret": "docker_username" },
        password: { "from_secret": "docker_password" },
        spec: "docker/" + binary + "/manifest.tmpl",
        ignore_missing: true,
      },
    },
    {
      name: "microbadger",
      image: "plugins/webhook:1",
      pull: "always",
      settings: {
        url: { "from_secret": "microbadger_docker" },
      },
    },
  ],
  depends_on: [
    "linux-amd64-" + binary,
    "linux-arm64-" + binary,
    "linux-arm-" + binary,
  ],
  trigger: {
    branch: [ "master" ],
    event: [ "push", "tag" ],
  },
};

[
  PipelineTesting,
  PipelineBuild("docker", "linux", "amd64"),
  PipelineBuild("docker", "linux", "arm64"),
  PipelineBuild("docker", "linux", "arm"),
  PipelineBuild("gcr", "linux", "amd64"),
  PipelineBuild("gcr", "linux", "arm64"),
  PipelineBuild("gcr", "linux", "arm"),
  PipelineBuild("ecr", "linux", "amd64"),
  PipelineBuild("ecr", "linux", "arm64"),
  PipelineBuild("ecr", "linux", "arm"),
  PipelineBuild("heroku", "linux", "amd64"),
  PipelineBuild("heroku", "linux", "arm64"),
  PipelineBuild("heroku", "linux", "arm"),
  PipelineNotifications("docker"),
  PipelineNotifications("gcr"),
  PipelineNotifications("ecr"),
  PipelineNotifications("heroku"),
]
