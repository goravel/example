package feature

import (
	"testing"

	"github.com/goravel/framework/contracts/binding"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
	"goravel/tests/telemetry"
)

// plainServiceName isolates this suite's telemetry in the shared backends;
// every telemetry suite must query by its own service name.
const plainServiceName = "goravel-plain"

type TelemetryTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestTelemetryTestSuite(t *testing.T) {
	suite.Run(t, &TelemetryTestSuite{})
}

func (s *TelemetryTestSuite) SetupSuite() {
	telemetry.EnsureStack(s.T(), telemetry.ServiceJaeger, telemetry.ServicePrometheus, telemetry.ServiceLoki, telemetry.ServiceCollector)

	scope, err := tests.OverrideConfig(map[string]any{
		"telemetry.service.name": plainServiceName,
	})
	// Cleanup instead of TearDownSuite: it restores the config even when a
	// later assertion in SetupSuite fails.
	s.T().Cleanup(func() {
		s.NoError(scope.Restore())
	})
	s.Require().NoError(err)

	// The /telemetry handler runs an automatically instrumented DB query. The DB
	// connection and its bound tracer are cached process-wide and survive
	// App().Restart(), so a connection an earlier suite built keeps that suite's
	// service name. Orm().Fresh() clears the connection cache and
	// App().Fresh(binding.DB) drops the resolved DB facade, so the next query
	// rebuilds the instrument under plainServiceName. Then recreate the users
	// table the query reads.
	facades.Orm().Fresh()
	facades.App().Fresh(binding.DB)
	s.RefreshDatabase()

	resp, err := s.Http(s.T()).Get("/telemetry")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	// An empty user id drives the validation error path so the result="error"
	// metric and the errored spans are emitted as well.
	errResp, err := s.Http(s.T()).Get("/telemetry?user=")
	s.Require().NoError(err)
	errResp.AssertSuccessful()
}

func (s *TelemetryTestSuite) TestTraces() {
	telemetry.AwaitTraces(s.T(), plainServiceName,
		"GET /telemetry", "HTTP GET", "user.UserService/GetUser", "GET /grpc/user",
		"users.process", "users.consume", "SELECT users")
}

func (s *TelemetryTestSuite) TestMetrics() {
	telemetry.AwaitMetric(s.T(), `grpc_controller_total{service_name="`+plainServiceName+`"}`,
		"grpc_controller_total", "GrpcController/User")
	telemetry.AwaitMetric(s.T(), `users_processed_total{service_name="`+plainServiceName+`",result="ok"}`,
		"users_processed_total")
	telemetry.AwaitMetric(s.T(), `users_processed_total{service_name="`+plainServiceName+`",result="error"}`,
		"users_processed_total")
}

func (s *TelemetryTestSuite) TestLogs() {
	telemetry.AwaitLogs(s.T(), `{service_name="`+plainServiceName+`"}`,
		"test telemetry log")
}
