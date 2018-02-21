package main

import "testing"

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
