package docker

import (
	"encoding/pem"
	"fmt"
)

// decodePEMCertificates extracts DER-encoded CERTIFICATE blocks from a PEM
// (or PEM bundle). Non-certificate blocks are skipped. Returns an error when
// no certificate blocks are present.
func decodePEMCertificates(pemBytes []byte) ([][]byte, error) {
	var ders [][]byte
	rest := pemBytes
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		ders = append(ders, block.Bytes)
	}
	if len(ders) == 0 {
		return nil, fmt.Errorf("no CERTIFICATE PEM blocks found")
	}
	return ders, nil
}
