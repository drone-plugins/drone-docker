//go:build !linux && !windows
// +build !linux,!windows

package docker

import "fmt"

// installHarnessCA is a no-op on platforms where egress control is not yet
// supported (e.g. macOS). HARNESS_CA_PATH is honored on Linux and Windows.
func installHarnessCA(caPEM []byte) {
	fmt.Println("HARNESS_CA_PATH is set but egress-proxy CA trust injection is not supported on this platform; skipping.")
}
