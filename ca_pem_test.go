package docker

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// Real self-signed CAs generated for these tests (CN=HarnessTestCA / HarnessTestCA2).
const harnessTestCAPEM = `-----BEGIN CERTIFICATE-----
MIIDETCCAfmgAwIBAgIUUnMVT6JwuvsR8lPanhZxYJnqUPYwDQYJKoZIhvcNAQEL
BQAwGDEWMBQGA1UEAwwNSGFybmVzc1Rlc3RDQTAeFw0yNjA3MTUwNzMxMDBaFw0z
NjA3MTIwNzMxMDBaMBgxFjAUBgNVBAMMDUhhcm5lc3NUZXN0Q0EwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDJOssgNxmGP8itq7IxObLY2igTK3/YeMIJ
9gvjWzUGe/RRrWqWJvfplKScNaIbyGaABYbDaGURnL3uGr5QQWEUbqvOMlEYaoNq
21TmVbEl6Sponc+4aFfDPP9CFAOn2tqyrZaU3wZIjSUWbGo2wwseWULzaQxZ8lZ6
a+kgjipymolfTZeHbJQAk/hwXVhbrz6xBbOhue/K/tFJxFBuawvwgXEi+2ywzXgy
fn80wYWdjeFEpFAtNKyjg+bu36DtHikwDS9XqZzgSZnGhDKEGh3cj0hd3/nhr59h
FCDm56XsBvWQVHmNmiuuxCn3svzSa9qdyk6tcIqZN8wKz1J7lFRfAgMBAAGjUzBR
MB0GA1UdDgQWBBSrLR7fG/bwSozrDipyfT5MtiuYjTAfBgNVHSMEGDAWgBSrLR7f
G/bwSozrDipyfT5MtiuYjTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUA
A4IBAQAq9+CxKqe2vhBTz+5i9NBpxzxFy3fZTKcakuB98hafnNw29D4VggjV7sgo
99En87gkTIPsiOsRa0YH3wl5aa5EUp/Dx0ZfXasjzS5BFSZKZDYcuLuuUEeK9N6j
UBxaWDxNRjENevkX8SWpbxyNkcDGZM2IfJvVA/g7h8ERZLjlu9iOTqb6jt3PDe4q
CNoMbU5L3md3gk8mfoEZAX/IIUmbw7hAy7zE9dEfbKHeXl3FrcXPL0WmLePcK9BF
8XEpsOUSoq4sMPdVxaPF7l+SKDbdsR4gKtwlEMLefiiLqk0k4HhygjzzRgdsMOc3
S8dL7zv0BrAFy//XsfmsVeVFVhRN
-----END CERTIFICATE-----
`

const harnessTestCA2PEM = `-----BEGIN CERTIFICATE-----
MIIDEzCCAfugAwIBAgIUR5AfE7MTe0ucAk1JzlqvOJd3Z7AwDQYJKoZIhvcNAQEL
BQAwGTEXMBUGA1UEAwwOSGFybmVzc1Rlc3RDQTIwHhcNMjYwNzE1MDczMTAwWhcN
MzYwNzEyMDczMTAwWjAZMRcwFQYDVQQDDA5IYXJuZXNzVGVzdENBMjCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAI2EDM2Ac+jwhP8OaWFFb+DlVwzKSkEI
LoaZ41RQGHZaCckTLtaCtIBaNQHLWqj8q6I0fIz+8mmaXuxRKDjnhGl3AFIJ2wql
Hbyywo2MQcu+6wfs2Ewd887NDb5wrNJs/gl1GYRilSAd8n53T5IJUHHA0IGxtq4g
9XwS6OTmfBdQ3ycEsi9Yexd5Oz79sLMDhBnt21nYWrEO8Kumgyfd7gfQUBcGc1nH
NhoBrZffLmo+cljGqPNEFyAlWEnsmiqZnYh/EOUXGdXxsZMvyMWt02kLe0BlGmJr
PrI8iJCMElj86UhKK2oIjWLMA6cviYSkw9jyiihhodSUChhmqOABoe8CAwEAAaNT
MFEwHQYDVR0OBBYEFKK9RP95NSP4xmxN+hojPXlxVVatMB8GA1UdIwQYMBaAFKK9
RP95NSP4xmxN+hojPXlxVVatMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
BQADggEBABlrORwXmhmO+eIdTHZMuad72JMEJeHxiRCGbwH6vAVb03YZAXLmIf31
sBKkUCDV+Vm0MaTYmYncK3eZZBn6teEWcx7+L9XlF7ckVKBqj9xkwFUdfVipCimD
cjBl/H+2PEY5Hzoy/uf2/sE1czrfHA2HeEIEacVyht9t0HTtqXQqrn44ijESkNtb
SqqaVYCcceS85ID2oJYYKQycMnFJd0VwSHlD+Lu4ZubXbrg+pzSqw6olTOwCE7Q2
LhLxQRB4f8Xw807aIuFK2mKbPs+4dmPkhClM3rbFzzldQJ4lKLnvZHcFtXJUXg2a
UzwP8JQn8xuz7gIZxMO3vGqTweP48yc=
-----END CERTIFICATE-----
`

func TestDecodePEMCertificates_Single(t *testing.T) {
	ders, err := decodePEMCertificates([]byte(harnessTestCAPEM))
	if err != nil {
		t.Fatalf("decodePEMCertificates: %v", err)
	}
	if len(ders) != 1 {
		t.Fatalf("got %d certs, want 1", len(ders))
	}
	cert, err := x509.ParseCertificate(ders[0])
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}
	if cert.Subject.CommonName != "HarnessTestCA" {
		t.Fatalf("CN = %q, want HarnessTestCA", cert.Subject.CommonName)
	}
}

func TestDecodePEMCertificates_Bundle(t *testing.T) {
	bundle := harnessTestCAPEM + "\n" + harnessTestCA2PEM
	ders, err := decodePEMCertificates([]byte(bundle))
	if err != nil {
		t.Fatalf("decodePEMCertificates: %v", err)
	}
	if len(ders) != 2 {
		t.Fatalf("got %d certs, want 2", len(ders))
	}
	for i, der := range ders {
		if _, err := x509.ParseCertificate(der); err != nil {
			t.Fatalf("cert %d: ParseCertificate: %v", i, err)
		}
	}
}

func TestDecodePEMCertificates_SkipsNonCertificateBlocks(t *testing.T) {
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not-a-key")})
	mixed := append(append([]byte{}, keyPEM...), []byte(harnessTestCAPEM)...)
	ders, err := decodePEMCertificates(mixed)
	if err != nil {
		t.Fatalf("decodePEMCertificates: %v", err)
	}
	if len(ders) != 1 {
		t.Fatalf("got %d certs, want 1 (private key block skipped)", len(ders))
	}
}

func TestDecodePEMCertificates_Empty(t *testing.T) {
	_, err := decodePEMCertificates([]byte("not a pem\n"))
	if err == nil {
		t.Fatal("expected error for input with no CERTIFICATE blocks")
	}
}

func TestDecodePEMCertificates_MatchesOpenSSLDER(t *testing.T) {
	ders, err := decodePEMCertificates([]byte(harnessTestCAPEM))
	if err != nil {
		t.Fatal(err)
	}
	origBlock, _ := pem.Decode([]byte(harnessTestCAPEM))
	if !bytes.Equal(origBlock.Bytes, ders[0]) {
		t.Fatal("decoded DER does not match original PEM payload")
	}
}
