package telemetry

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

var certsDir string

// GenerateCollectorCerts writes a throwaway CA and a localhost server
// certificate for the TLS collector into dir and returns the CA path. The app
// trusts the CA via "tls.ca"; the collector serves with the server pair
// mounted by docker compose. Call it before EnsureStack starts the TLS
// collector, which mounts the directory at startup. TeardownStack removes the
// directory.
func GenerateCollectorCerts(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	certsDir = dir

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "goravel example test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return "", err
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return "", err
	}

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return "", err
	}
	serverKeyDER, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		return "", err
	}

	caPath := filepath.Join(dir, "ca.crt")
	if err := writePEM(caPath, "CERTIFICATE", caDER); err != nil {
		return "", err
	}
	if err := writePEM(filepath.Join(dir, "server.crt"), "CERTIFICATE", serverDER); err != nil {
		return "", err
	}
	if err := writePEM(filepath.Join(dir, "server.key"), "EC PRIVATE KEY", serverKeyDER); err != nil {
		return "", err
	}

	return caPath, nil
}

// writePEM uses 0644 so the non-root collector container can read the mounted
// files; these are throwaway test credentials.
func writePEM(path, blockType string, der []byte) error {
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), 0o644)
}
