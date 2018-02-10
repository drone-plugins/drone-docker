package main

import "testing"

func TestGetRepoName(t *testing.T) {
	// map full repo path to expected repo name
	repos := map[string]string{
		"000000000000.dkr.ecr.us-east-1.amazonaws.com/repo":                     "repo",
		"000000000000.dkr.ecr.us-east-1.amazonaws.com/namespace/repo":           "namespace/repo",
		"000000000000.dkr.ecr.us-east-1.amazonaws.com/namespace/namespace/repo": "namespace/namespace/repo",
	}

	for repo, name := range repos {
		splitName := getRepoName(repo)
		if splitName != name {
			t.Errorf("%s is not equal to %s.", splitName, name)
		}
	}
}
