package feature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

// tlsServiceName isolates this suite's telemetry from TelemetryTestSuite,
// which exports through the plaintext collector under the app name.
const tlsServiceName = "goravel-tls"

type TelemetryTLSTestSuite struct {
	suite.Suite
	tests.TestCase
	originalConfig map[string]any
}

func TestTelemetryTLSTestSuite(t *testing.T) {
	suite.Run(t, &TelemetryTLSTestSuite{})
}

func (s *TelemetryTLSTestSuite) SetupSuite() {
	certsDir, err := filepath.Abs("../../tests/testdata/otel-tls")
	s.Require().NoError(err)

	caPath, err := generateCollectorCerts(certsDir)
	s.Require().NoError(err)

	s.False(facades.Process().Path("../../").Run("docker compose up -d prometheus jaeger loki otel-collector-tls").Failed())
	time.Sleep(5 * time.Second)

	s.originalConfig = map[string]any{
		"telemetry.service.name": facades.Config().GetString("telemetry.service.name"),
	}
	// Signal-specific OTLP endpoints carry the full URL: the SDK uses them
	// as-is, without appending the default /v1/<signal> path.
	endpoints := map[string]string{
		"otlptrace":  "https://localhost:4319/v1/traces",
		"otlpmetric": "https://localhost:4319/v1/metrics",
		"otlplog":    "https://localhost:4319/v1/logs",
	}
	for exporter, endpoint := range endpoints {
		prefix := "telemetry.exporters." + exporter
		s.originalConfig[prefix+".endpoint"] = facades.Config().GetString(prefix + ".endpoint")
		s.originalConfig[prefix+".insecure"] = facades.Config().GetBool(prefix + ".insecure")
		s.originalConfig[prefix+".tls.ca"] = facades.Config().GetString(prefix + ".tls.ca")

		facades.Config().Add(prefix+".endpoint", endpoint)
		facades.Config().Add(prefix+".insecure", false)
		facades.Config().Add(prefix+".tls.ca", caPath)
	}
	facades.Config().Add("telemetry.service.name", tlsServiceName)

	s.Require().NoError(facades.App().Restart())

	resp, err := s.Http(s.T()).Get("/telemetry")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	// Wait for telemetry data to be exported
	time.Sleep(11 * time.Second)
}

func (s *TelemetryTLSTestSuite) TearDownSuite() {
	for key, value := range s.originalConfig {
		facades.Config().Add(key, value)
	}
	s.NoError(facades.App().Restart())

	s.False(facades.Process().Path("../../").Run("docker compose down").Failed())
}

func (s *TelemetryTLSTestSuite) TestTraces() {
	resp, err := facades.Http().Get("http://localhost:16686/api/traces?service=" + tlsServiceName)
	s.NoError(err)

	body, err := resp.Body()
	s.NoError(err)

	s.Contains(body, "GET /telemetry")
	s.Contains(body, "user.UserService/GetUser")
}

func (s *TelemetryTLSTestSuite) TestMetrics() {
	query := url.QueryEscape(`grpc_controller_total{service_name="` + tlsServiceName + `"}`)
	resp, err := facades.Http().Get("http://localhost:9090/api/v1/query?query=" + query)
	s.NoError(err)

	body, err := resp.Body()
	s.NoError(err)

	s.Contains(body, "grpc_controller_total")
	s.Contains(body, tlsServiceName)
}

// Logs are not asserted here: the otel log channel caches its handler
// process-wide, so after TelemetryTestSuite has used it once, records
// emitted after the restart in SetupSuite no longer reach the exporter.
// The TLS transport itself is shared by all three signals and is covered
// by the trace and metric assertions above.

// generateCollectorCerts writes a throwaway CA and a localhost server
// certificate for the TLS collector into dir and returns the CA path.
// The app trusts the CA via "tls.ca"; the collector serves with the
// server pair mounted by docker compose.
func generateCollectorCerts(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

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

// writePEM uses 0644 so the non-root collector container can read the
// mounted files; these are throwaway test credentials.
func writePEM(path, blockType string, der []byte) error {
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), 0o644)
}
