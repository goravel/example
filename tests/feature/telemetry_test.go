package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
	"goravel/tests/telemetry"
)

// plainServiceName isolates this suite's telemetry in the shared backends;
// every telemetry suite must query by its own service name.
const plainServiceName = "goravel-plain"

type TelemetryTestSuite struct {
	suite.Suite
	tests.TestCase
	scope *telemetry.ConfigScope
}

func TestTelemetryTestSuite(t *testing.T) {
	suite.Run(t, &TelemetryTestSuite{})
}

func (s *TelemetryTestSuite) SetupSuite() {
	telemetry.EnsureStack(s.T(), telemetry.ServiceJaeger, telemetry.ServicePrometheus, telemetry.ServiceLoki, telemetry.ServiceCollector)

	scope, err := telemetry.OverrideConfig(map[string]any{
		"telemetry.service.name": plainServiceName,
	})
	s.scope = scope
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/telemetry")
	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *TelemetryTestSuite) TearDownSuite() {
	s.NoError(s.scope.Restore())
}

func (s *TelemetryTestSuite) TestTraces() {
	telemetry.AwaitTraces(s.T(), plainServiceName,
		"GET /telemetry", "HTTP GET", "user.UserService/GetUser", "GET /grpc/user")
}

func (s *TelemetryTestSuite) TestMetrics() {
	telemetry.AwaitMetric(s.T(), `grpc_controller_total{service_name="`+plainServiceName+`"}`,
		"grpc_controller_total", "GrpcController/User")
}

func (s *TelemetryTestSuite) TestLogs() {
	telemetry.AwaitLogs(s.T(), `{service_name="`+plainServiceName+`"}`,
		"test telemetry log")
}
