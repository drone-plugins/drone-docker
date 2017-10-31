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

func Test_defaultTags(t *testing.T) {
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
