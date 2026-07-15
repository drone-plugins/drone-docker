//go:build windows
// +build windows

package docker

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// installHarnessCA imports the Harness egress-proxy CA into the Windows
// certificate store under LocalMachine\Root (the machine-wide trusted roots the
// Docker daemon and Go's crypto/x509 consult). Best-effort: failures are logged,
// never fatal.
//
// Uses CryptoAPI (crypt32.dll) directly so it works on Nano Server images that
// lack certutil and PowerShell. CERT_STORE_ADD_REPLACE_EXISTING keeps the step
// idempotent across retries.
func installHarnessCA(caPEM []byte) {
	ders, err := decodePEMCertificates(caPEM)
	if err != nil {
		fmt.Printf("Could not decode Harness CA PEM: %s\n", err)
		fmt.Println("Base-image pulls through the egress proxy may fail with x509 errors.")
		return
	}

	storeName, err := windows.UTF16PtrFromString("ROOT")
	if err != nil {
		fmt.Printf("Could not prepare Windows root store name: %s\n", err)
		return
	}

	store, err := windows.CertOpenStore(
		windows.CERT_STORE_PROV_SYSTEM,
		0,
		0,
		windows.CERT_SYSTEM_STORE_LOCAL_MACHINE,
		uintptr(unsafe.Pointer(storeName)),
	)
	if err != nil {
		fmt.Printf("Could not open LocalMachine\\Root store: %s\n", err)
		fmt.Println("Base-image pulls through the egress proxy may fail with x509 errors.")
		return
	}
	defer windows.CertCloseStore(store, 0)

	for i, der := range ders {
		if err := addCertToStore(store, der); err != nil {
			fmt.Printf("Warning: failed to add Harness CA cert %d to LocalMachine\\Root: %s\n", i+1, err)
			fmt.Println("Base-image pulls through the egress proxy may fail with x509 errors.")
			return
		}
	}

	fmt.Printf("Installed %d Harness egress CA certificate(s) into the Windows LocalMachine\\Root store.\n", len(ders))
}

func addCertToStore(store windows.Handle, der []byte) error {
	if len(der) == 0 {
		return fmt.Errorf("empty DER certificate")
	}
	ctx, err := windows.CertCreateCertificateContext(
		windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING,
		&der[0],
		uint32(len(der)),
	)
	if err != nil {
		return fmt.Errorf("CertCreateCertificateContext: %w", err)
	}
	defer windows.CertFreeCertificateContext(ctx)

	if err := windows.CertAddCertificateContextToStore(
		store,
		ctx,
		windows.CERT_STORE_ADD_REPLACE_EXISTING,
		nil,
	); err != nil {
		return fmt.Errorf("CertAddCertificateContextToStore: %w", err)
	}
	return nil
}
