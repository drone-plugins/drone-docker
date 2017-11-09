package docker

import (
	"reflect"
	"testing"
)

func Test_stripTagPrefix(t *testing.T) {
	var tests = []struct {
		Before string
		After  string
	}{
		{"refs/tags/1.0.0", "1.0.0"},
		{"refs/tags/v1.0.0", "1.0.0"},
		{"v1.0.0", "1.0.0"},
	}

	for _, test := range tests {
		got, want := stripTagPrefix(test.Before), test.After
		if got != want {
			t.Errorf("Got tag %s, want %s", got, want)
		}
	}
}

func TestDefaultTags(t *testing.T) {
	var tests = []struct {
		Before        string
		DefaultBranch string
		After         []string
	}{
		{"", "master", []string{"latest"}},
		{"refs/heads/master", "master", []string{"latest"}},
		{"refs/tags/0.9.0", "master", []string{"0.9", "0.9.0"}},
		{"refs/tags/1.0.0", "master", []string{"1", "1.0", "1.0.0"}},
		{"refs/tags/v1.0.0", "master", []string{"1", "1.0", "1.0.0"}},
		{"refs/tags/v1.0.0-alpha.1", "master", []string{"1.0.0-alpha.1"}},

		// malformed or errors
		{"refs/tags/x1.0.0", "master", []string{"latest"}},
		{"v1.0.0", "master", []string{"latest"}},

		// defualt branch
		{"refs/heads/master", "master", []string{"latest"}},
		{"refs/heads/test", "master", []string{}},
		{"refs/tags/v1.0.0", "master", []string{"1", "1.0", "1.0.0"}},
	}

	for _, test := range tests {
		got, want := DefaultTags(test.Before, test.DefaultBranch), test.After
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got tag %v, want %v", got, want)
		}
	}
}

func TestDefaultTagSuffix(t *testing.T) {
	var tests = []struct {
		Before        string
		DefaultBranch string
		Suffix        string
		After         []string
	}{
		// without suffix
		{
			After:         []string{"latest"},
			DefaultBranch: "",
		},
		{
			Before:        "refs/tags/v1.0.0",
			DefaultBranch: "",
			After: []string{
				"1",
				"1.0",
				"1.0.0",
			},
		},
		// with suffix
		{
			DefaultBranch: "",
			Suffix:        "linux-amd64",
			After:         []string{"linux-amd64"},
		},
		{
			DefaultBranch: "",
			Before:        "refs/tags/v1.0.0",
			Suffix:        "linux-amd64",
			After: []string{
				"1-linux-amd64",
				"1.0-linux-amd64",
				"1.0.0-linux-amd64",
			},
		},
		{
			DefaultBranch: "",
			Suffix:        "nanoserver",
			After:         []string{"nanoserver"},
		},
		{
			DefaultBranch: "",
			Before:        "refs/tags/v1.9.2",
			Suffix:        "nanoserver",
			After: []string{
				"1-nanoserver",
				"1.9-nanoserver",
				"1.9.2-nanoserver",
			},
		},
	}

	for _, test := range tests {
		got, want := DefaultTagSuffix(test.Before, test.Suffix, test.DefaultBranch), test.After
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got tag %v, want %v", got, want)
		}
	}
}

func Test_stripHeadPrefix(t *testing.T) {
	type args struct {
		ref string
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args: args{
				ref: "refs/heads/master",
			},
			want: "master",
		},
	}
	for _, tt := range tests {
		if got := stripHeadPrefix(tt.args.ref); got != tt.want {
			t.Errorf("stripHeadPrefix() = %v, want %v", got, tt.want)
		}
	}
}
