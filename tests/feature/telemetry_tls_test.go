package feature

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
	"goravel/tests/telemetry"
)

// tlsServiceName isolates this suite's telemetry in the shared backends;
// every telemetry suite must query by its own service name.
const tlsServiceName = "goravel-tls"

type TelemetryTLSTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestTelemetryTLSTestSuite(t *testing.T) {
	suite.Run(t, &TelemetryTLSTestSuite{})
}

func (s *TelemetryTLSTestSuite) SetupSuite() {
	certsDir, err := filepath.Abs("../testdata/otel-tls")
	s.Require().NoError(err)

	caPath, err := telemetry.GenerateCollectorCerts(certsDir)
	s.Require().NoError(err)

	telemetry.EnsureStack(s.T(), telemetry.ServiceJaeger, telemetry.ServicePrometheus, telemetry.ServiceLoki, telemetry.ServiceCollectorTLS)

	overrides := map[string]any{
		"telemetry.service.name": tlsServiceName,
	}
	// Signal-specific OTLP endpoints carry the full URL: the SDK uses them
	// as-is, without appending the default /v1/<signal> path.
	for exporter, endpoint := range map[string]string{
		"otlptrace":  "https://localhost:4319/v1/traces",
		"otlpmetric": "https://localhost:4319/v1/metrics",
		"otlplog":    "https://localhost:4319/v1/logs",
	} {
		prefix := "telemetry.exporters." + exporter
		overrides[prefix+".endpoint"] = endpoint
		overrides[prefix+".insecure"] = false
		overrides[prefix+".tls.ca"] = caPath
	}

	scope, err := telemetry.OverrideConfig(overrides)
	// Cleanup instead of TearDownSuite: it restores the config even when a
	// later assertion in SetupSuite fails.
	s.T().Cleanup(func() {
		s.NoError(scope.Restore())
	})
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/telemetry")
	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *TelemetryTLSTestSuite) TestTraces() {
	telemetry.AwaitTraces(s.T(), tlsServiceName,
		"GET /telemetry", "user.UserService/GetUser")
}

func (s *TelemetryTLSTestSuite) TestMetrics() {
	telemetry.AwaitMetric(s.T(), `grpc_controller_total{service_name="`+tlsServiceName+`"}`,
		"grpc_controller_total", "GrpcController/User")
}

func (s *TelemetryTLSTestSuite) TestLogs() {
	telemetry.AwaitLogs(s.T(), `{service_name="`+tlsServiceName+`"}`,
		"test telemetry log")
}
