package docker

import (
	"bytes"
	"os/exec"
	"testing"
)

func Test_trace(t *testing.T) {
	tests := []struct {
		name    string
		cmd    *exec.Cmd
		want string
	}{
		{"no args", exec.Command("docker", "system", "prune", "-f"), "+ docker system prune -f\n"},
		{"with args", exec.Command("docker", "build", "--build-arg", "SECRET=e7fa7552-e08a-4479-9959-7169df4ce686"), "+ docker build --build-arg SECRET={redacted}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			trace(out, tt.cmd)
			if gotOut := out.String(); gotOut != tt.want {
				t.Errorf("trace() = %v, want %v", gotOut, tt.want)
			}
		})
	}
}