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
		Before string
		After  []string
	}{
		{"", []string{"latest"}},
		{"refs/heads/master", []string{"latest"}},
		{"refs/tags/0.9.0", []string{"0.9", "0.9.0"}},
		{"refs/tags/1.0.0", []string{"1", "1.0", "1.0.0"}},
		{"refs/tags/v1.0.0", []string{"1", "1.0", "1.0.0"}},
		{"refs/tags/v1.0.0-alpha.1", []string{"1.0.0-alpha.1"}},

		// malformed or errors
		{"refs/tags/x1.0.0", []string{"latest"}},
		{"v1.0.0", []string{"latest"}},
	}

	for _, test := range tests {
		got, want := DefaultTags(test.Before), test.After
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got tag %v, want %v", got, want)
		}
	}
}

func TestDefaultTagSuffix(t *testing.T) {
	var tests = []struct {
		Before string
		Suffix string
		After  []string
	}{
		// without suffix
		{
			After: []string{"latest"},
		},
		{
			Before: "refs/tags/v1.0.0",
			After: []string{
				"1",
				"1.0",
				"1.0.0",
			},
		},
		// with suffix
		{
			Suffix: "linux-amd64",
			After:  []string{"linux-amd64"},
		},
		{
			Before: "refs/tags/v1.0.0",
			Suffix: "linux-amd64",
			After: []string{
				"1-linux-amd64",
				"1.0-linux-amd64",
				"1.0.0-linux-amd64",
			},
		},
		{
			Suffix: "nanoserver",
			After:  []string{"nanoserver"},
		},
		{
			Before: "refs/tags/v1.9.2",
			Suffix: "nanoserver",
			After: []string{
				"1-nanoserver",
				"1.9-nanoserver",
				"1.9.2-nanoserver",
			},
		},
		{
			Before: "refs/tags/v18.06.0",
			Suffix: "nanoserver",
			After: []string{
				"18-nanoserver",
				"18.06-nanoserver",
				"18.06.0-nanoserver",
			},
		},
	}

	for _, test := range tests {
		got, want := DefaultTagSuffix(test.Before, test.Suffix), test.After
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

func TestUseDefaultTag(t *testing.T) {
	type args struct {
		ref           string
		defaultBranch string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "latest tag for default branch",
			args: args{
				ref:           "refs/heads/master",
				defaultBranch: "master",
			},
			want: true,
		},
		{
			name: "build from tags",
			args: args{
				ref:           "refs/tags/v1.0.0",
				defaultBranch: "master",
			},
			want: true,
		},
		{
			name: "skip build for not default branch",
			args: args{
				ref:           "refs/heads/develop",
				defaultBranch: "master",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := UseDefaultTag(tt.args.ref, tt.args.defaultBranch); got != tt.want {
			t.Errorf("%q. UseDefaultTag() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
