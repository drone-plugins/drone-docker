//go:build linux
// +build linux

package docker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// anchorCandidate is a distro CA anchor directory plus the tool that refreshes
// the consolidated bundle from it.
type anchorCandidate struct {
	dir     string
	refresh []string
}

// linuxAnchorCandidates and linuxBundleCandidates are package vars (not consts)
// so tests can point them at a temp trust store. Order matters: the first
// existing entry wins for the anchor install.
var linuxAnchorCandidates = []anchorCandidate{
	{"/usr/local/share/ca-certificates", []string{"update-ca-certificates"}},     // Debian/Ubuntu/Alpine
	{"/etc/pki/ca-trust/source/anchors", []string{"update-ca-trust", "extract"}}, // RHEL/CentOS/Fedora
}

var linuxBundleCandidates = []string{
	"/etc/ssl/certs/ca-certificates.crt", // Debian/Ubuntu/Alpine
	"/etc/pki/tls/certs/ca-bundle.crt",   // RHEL/CentOS
}

// installHarnessCA drops the CA into the distro anchor directory and refreshes
// the system bundle. It also appends directly to the consolidated bundle so
// trust is effective even on images that lack update-ca-certificates (e.g.
// minimal distros); the Docker daemon and Go's crypto/x509 read from that
// bundle. Best-effort: failures are logged, never fatal.
func installHarnessCA(caPEM []byte) {
	anchorCandidates := linuxAnchorCandidates

	installedViaAnchor := false
	for _, c := range anchorCandidates {
		if _, err := os.Stat(c.dir); err != nil {
			continue
		}
		dst := filepath.Join(c.dir, "harness-egress-ca.crt")
		if err := os.WriteFile(dst, caPEM, 0644); err != nil {
			fmt.Printf("Could not write Harness CA to %s: %s\n", dst, err)
			continue
		}
		if _, err := exec.LookPath(c.refresh[0]); err != nil {
			continue
		}
		cmd := exec.Command(c.refresh[0], c.refresh[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: %s failed: %s\n", c.refresh[0], err)
			continue
		}
		fmt.Printf("Installed Harness egress CA into %s and refreshed system trust store.\n", c.dir)
		installedViaAnchor = true
		break
	}

	// Belt-and-suspenders: append to the consolidated bundle(s) directly so the
	// CA is trusted even when no refresh tool exists or the anchor step failed.
	bundleCandidates := linuxBundleCandidates
	appended := false
	for _, bundle := range bundleCandidates {
		if _, err := os.Stat(bundle); err != nil {
			continue
		}
		if bundleContainsCA(bundle, caPEM) {
			appended = true
			continue
		}
		f, err := os.OpenFile(bundle, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Could not append Harness CA to %s: %s\n", bundle, err)
			continue
		}
		payload := caPEM
		if !bytes.HasPrefix(payload, []byte("\n")) {
			payload = append([]byte("\n"), payload...)
		}
		if _, err := f.Write(payload); err != nil {
			fmt.Printf("Could not append Harness CA to %s: %s\n", bundle, err)
		} else {
			fmt.Printf("Appended Harness egress CA to %s.\n", bundle)
			appended = true
		}
		f.Close()
	}

	if !installedViaAnchor && !appended {
		fmt.Println("Warning: could not install the Harness egress CA into any known trust store location; base-image pulls through the egress proxy may fail with x509 errors.")
	}
}

// bundleContainsCA reports whether the CA bytes are already present in the
// bundle, to keep the step idempotent across retries.
func bundleContainsCA(bundle string, caPEM []byte) bool {
	existing, err := os.ReadFile(bundle)
	if err != nil {
		return false
	}
	return bytes.Contains(existing, bytes.TrimSpace(caPEM))
}
