//go:build windows
// +build windows

package docker

import (
	"fmt"
	"os"
	"os/exec"
)

// installHarnessCA imports the Harness egress-proxy CA into the Windows
// certificate store under LocalMachine\Root (the machine-wide trusted roots the
// Docker daemon and Go's crypto/x509 consult). Best-effort: failures are logged,
// never fatal.
//
// certutil needs a file path, so the (already-validated) CA bytes are written to
// a temp .crt first, then added with `certutil -addstore -f Root <file>`.
func installHarnessCA(caPEM []byte) {
	tmp, err := os.CreateTemp("", "harness-egress-ca-*.crt")
	if err != nil {
		fmt.Printf("Could not create temp file for Harness CA: %s\n", err)
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(caPEM); err != nil {
		tmp.Close()
		fmt.Printf("Could not write temp Harness CA: %s\n", err)
		return
	}
	tmp.Close()

	if _, err := exec.LookPath("certutil"); err != nil {
		fmt.Printf("certutil not found on PATH; cannot install Harness egress CA: %s\n", err)
		fmt.Println("Base-image pulls through the egress proxy may fail with x509 errors.")
		return
	}

	// -f overwrites an existing entry, keeping the step idempotent across retries.
	cmd := exec.Command("certutil", "-addstore", "-f", "Root", tmpPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: certutil -addstore Root failed: %s\n", err)
		fmt.Println("Base-image pulls through the egress proxy may fail with x509 errors.")
		return
	}
	fmt.Println("Installed Harness egress CA into the Windows LocalMachine\\Root store.")
}
