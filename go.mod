module github.com/drone-plugins/drone-docker

require (
	github.com/aws/aws-sdk-go v1.16.15
	github.com/coreos/go-semver v0.2.0
	github.com/joho/godotenv v1.3.0
	github.com/sirupsen/logrus v1.3.0
	github.com/urfave/cli v1.20.0
	golang.org/x/net v0.0.0-20190213065845-3a22650c66bd // indirect
	golang.org/x/text v0.3.0 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace github.com/urfave/cli => github.com/bradrydzewski/cli v0.0.0-20190108225652-0d51abd87c77
