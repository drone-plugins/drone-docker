package docker

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestCommandStartDaemon(t *testing.T) {
	tcs := []struct {
		name   string
		daemon Daemon
		want   *exec.Cmd
	}{
		{
			name: "multi mirrors",
			daemon: Daemon{
				StoragePath: "/var/lib/docker",
				Mirror:      "https://a.com,https://b.com,https://c.com",
			},
			want: exec.Command(
				dockerdExe,
				"--data-root",
				"/var/lib/docker",
				"--host=unix:///var/run/docker.sock",
				"--registry-mirror",
				"https://a.com",
				"--registry-mirror",
				"https://b.com",
				"--registry-mirror",
				"https://c.com",
			),
		},
		{
			name: "single mirrors",
			daemon: Daemon{
				StoragePath: "/var/lib/docker",
				Mirror:      "https://a.com",
			},
			want: exec.Command(
				dockerdExe,
				"--data-root",
				"/var/lib/docker",
				"--host=unix:///var/run/docker.sock",
				"--registry-mirror",
				"https://a.com",
			),
		},
		{
			name: "zero mirrors",
			daemon: Daemon{
				StoragePath: "/var/lib/docker",
			},
			want: exec.Command(
				dockerdExe,
				"--data-root",
				"/var/lib/docker",
				"--host=unix:///var/run/docker.sock",
			),
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := commandDaemon(tc.daemon)

			if !reflect.DeepEqual(cmd.String(), tc.want.String()) {
				t.Errorf("Got cmd %v, want %v", cmd, tc.want)
			}
		})
	}
}
