package main

import "testing"

func TestNormalizeTags(t *testing.T) {
	type testCase struct {
		Input string
		Output string
	}

	inputs := []string{"test", "test-branch", "prefix/test"}
	expectedOutputs := []string{"test", "test-branch", "prefix-test"}

	outputs := normalizeTags(inputs)

	for i, output := range outputs {
		if output != expectedOutputs[i] {
			t.Fatalf("expected %s, got %s", expectedOutputs[i], output)
		}
	}

}
