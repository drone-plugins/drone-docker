package docker

import (
	"reflect"
	"strings"
	"testing"
)

func Test_commandBuild(t *testing.T) {
	var tests = []struct {
		Input  Build
		Output []string
	}{
		{
			// Able to build without name by tagging directly. See #229
			Build{
				Dockerfile: "Dockerfile",
				Context:    "./context",
				Tags:       []string{"2", "2.1", "2.1.9"},
				Repo:       "example/hello-world",
			},
			[]string{
				"/usr/local/bin/docker", "build",
				"--rm=true",
				"-f", "Dockerfile",
				"-t", "example/hello-world:2",
				"-t", "example/hello-world:2.1",
				"-t", "example/hello-world:2.1.9",
				"./context",
				"--label", "org.label-schema.schema-version=1.0",
				"--label", "org.label-schema.build-date=",
				"--label", "org.label-schema.vcs-ref=",
				"--label", "org.label-schema.vcs-url=",
			},
		},
	}

	for _, test := range tests {
		got, want := commandBuild(test.Input).Args, test.Output

		// Remove build date.
		for i, v := range got {
			if strings.HasPrefix(v, "org.label-schema.build-date=") {
				got[i] = "org.label-schema.build-date="
			}
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got command %v, want %v", got, want)
		}
	}
}
