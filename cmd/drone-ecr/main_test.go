package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
)

func TestTrimHostname(t *testing.T) {
	registry := "000000000000.dkr.ecr.us-east-1.amazonaws.com"
	// map full repo path to expected repo name
	repos := map[string]string{
		registry + "/repo":                     "repo",
		registry + "/namespace/repo":           "namespace/repo",
		registry + "/namespace/namespace/repo": "namespace/namespace/repo",
	}

	for repo, name := range repos {
		splitName := trimHostname(repo, registry)
		if splitName != name {
			t.Errorf("%s is not equal to %s.", splitName, name)
		}
	}
}

func TestGetTagMutabilityString(t *testing.T) {
	testCases := []struct {
		name         string
		tagImmutable bool
		expected     string
	}{
		{
			name:         "mutable",
			tagImmutable: false,
			expected:     ecr.ImageTagMutabilityMutable,
		},
		{
			name:         "immutable",
			tagImmutable: true,
			expected:     ecr.ImageTagMutabilityImmutable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := getTagMutabilityString(tc.tagImmutable)
			if actual != tc.expected {
				t.Errorf("expected: %s, actual: %s", tc.expected, actual)
			}
		})
	}
}
