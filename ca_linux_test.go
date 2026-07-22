//go:build linux
// +build linux

package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempTrustStore points the Linux anchor/bundle candidates at a temp dir so
// installHarnessCA can be exercised without touching the real system store.
func withTempTrustStore(t *testing.T) (anchorDir, bundlePath string) {
	t.Helper()
	dir := t.TempDir()
	anchorDir = filepath.Join(dir, "anchors")
	bundlePath = filepath.Join(dir, "ca-certificates.crt")
	if err := os.MkdirAll(anchorDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Seed an existing bundle so the append path is exercised.
	if err := os.WriteFile(bundlePath, []byte("# existing roots\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origAnchors := linuxAnchorCandidates
	origBundles := linuxBundleCandidates
	// No refresh tool on the anchor entry so the test doesn't shell out.
	linuxAnchorCandidates = []anchorCandidate{{dir: anchorDir, refresh: []string{"harness-nonexistent-refresh-tool"}}}
	linuxBundleCandidates = []string{bundlePath}
	t.Cleanup(func() {
		linuxAnchorCandidates = origAnchors
		linuxBundleCandidates = origBundles
	})
	return anchorDir, bundlePath
}

func TestInstallHarnessCA_Linux(t *testing.T) {
	anchorDir, bundlePath := withTempTrustStore(t)

	installHarnessCA([]byte(testCAPEM))

	// Anchor file should be written.
	anchorFile := filepath.Join(anchorDir, "harness-egress-ca.crt")
	got, err := os.ReadFile(anchorFile)
	if err != nil {
		t.Fatalf("expected anchor file at %s: %s", anchorFile, err)
	}
	if string(got) != testCAPEM {
		t.Fatalf("anchor file = %q, want %q", string(got), testCAPEM)
	}

	// Bundle should have the CA appended (refresh tool is absent, so the
	// direct-append fallback is what provides trust here).
	bundle, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bundle), strings.TrimSpace(testCAPEM)) {
		t.Fatalf("bundle %q does not contain appended CA", string(bundle))
	}
	if !strings.HasPrefix(string(bundle), "# existing roots") {
		t.Fatalf("bundle append clobbered existing roots: %q", string(bundle))
	}
}

func TestInstallHarnessCA_Linux_Idempotent(t *testing.T) {
	_, bundlePath := withTempTrustStore(t)

	installHarnessCA([]byte(testCAPEM))
	installHarnessCA([]byte(testCAPEM))

	bundle, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(bundle), strings.TrimSpace(testCAPEM)); n != 1 {
		t.Fatalf("expected CA to appear exactly once after two installs, got %d", n)
	}
}
