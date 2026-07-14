package docker

import (
	"os"
	"path/filepath"
	"testing"
)

const testCAPEM = `-----BEGIN CERTIFICATE-----
MIIBAg==
-----END CERTIFICATE-----
`

// withStubbedInstall replaces the platform installer with a capturing stub for
// the duration of a test, returning a pointer to the captured CA bytes and a
// pointer to a call counter.
func withStubbedInstall(t *testing.T) (captured *[]byte, calls *int) {
	t.Helper()
	captured = new([]byte)
	calls = new(int)
	orig := installHarnessCAFn
	installHarnessCAFn = func(caPEM []byte) {
		*calls++
		*captured = append([]byte(nil), caPEM...)
	}
	t.Cleanup(func() { installHarnessCAFn = orig })
	return captured, calls
}

func TestTrustHarnessCA_Unset(t *testing.T) {
	os.Unsetenv("HARNESS_CA_PATH")
	_, calls := withStubbedInstall(t)

	trustHarnessCA()

	if *calls != 0 {
		t.Fatalf("expected installer not to be called when HARNESS_CA_PATH is unset, got %d calls", *calls)
	}
}

func TestTrustHarnessCA_MissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.crt")
	os.Setenv("HARNESS_CA_PATH", missing)
	defer os.Unsetenv("HARNESS_CA_PATH")
	_, calls := withStubbedInstall(t)

	trustHarnessCA()

	if *calls != 0 {
		t.Fatalf("expected installer not to be called when CA file is missing, got %d calls", *calls)
	}
}

func TestTrustHarnessCA_EmptyFile(t *testing.T) {
	empty := filepath.Join(t.TempDir(), "empty.crt")
	if err := os.WriteFile(empty, []byte("   \n\t"), 0644); err != nil {
		t.Fatal(err)
	}
	os.Setenv("HARNESS_CA_PATH", empty)
	defer os.Unsetenv("HARNESS_CA_PATH")
	_, calls := withStubbedInstall(t)

	trustHarnessCA()

	if *calls != 0 {
		t.Fatalf("expected installer not to be called for whitespace-only CA, got %d calls", *calls)
	}
}

func TestTrustHarnessCA_ValidFile(t *testing.T) {
	valid := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(valid, []byte(testCAPEM), 0644); err != nil {
		t.Fatal(err)
	}
	os.Setenv("HARNESS_CA_PATH", valid)
	defer os.Unsetenv("HARNESS_CA_PATH")
	captured, calls := withStubbedInstall(t)

	trustHarnessCA()

	if *calls != 1 {
		t.Fatalf("expected installer to be called once for a valid CA, got %d calls", *calls)
	}
	if string(*captured) != testCAPEM {
		t.Fatalf("installer received %q, want %q", string(*captured), testCAPEM)
	}
}
