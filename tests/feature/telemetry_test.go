package feature

import (
	"fmt"
	"testing"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type TelemetryTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestTelemetryTestSuite(t *testing.T) {
	suite.Run(t, &TelemetryTestSuite{})
}

func (s *TelemetryTestSuite) SetupSuite() {
	s.False(facades.Process().Path("../../").Run("docker compose up -d prometheus jaeger loki otel-collector").Failed())
	time.Sleep(5 * time.Second)

	resp, err := s.Http(s.T()).Get("/telemetry")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	// Wait for telemetry data to be exported
	time.Sleep(2 * time.Second)
}

func (s *TelemetryTestSuite) TearDownSuite() {
	s.False(facades.Process().Path("../../").Run("docker compose down").Failed())
}

func (s *TelemetryTestSuite) TestTraces() {
	appName := facades.Config().GetString("app.name")
	resp, err := facades.Http().Get("http://localhost:16686/api/traces?service=" + appName)
	s.NoError(err)

	body, err := resp.Body()
	s.NoError(err)

	s.Contains(body, "GET /telemetry")
	s.Contains(body, "HTTP GET")
	s.Contains(body, "user.UserService/GetUser")
	s.Contains(body, "GET /grpc/user")
}

func (s *TelemetryTestSuite) TestMetrics() {
	resp, err := facades.Http().Get("http://localhost:9090/api/v1/query?query=grpc_controller_total")
	s.NoError(err)

	body, err := resp.Body()
	s.NoError(err)

	s.Contains(body, "grpc_controller_total")
	s.Contains(body, "GrpcController/User")
}

func (s *TelemetryTestSuite) TestLogs() {
	appName := facades.Config().GetString("app.name")

	// Calculate time range (last 5 minutes to now)
	end := time.Now().UnixNano()
	start := time.Now().Add(-5 * time.Minute).UnixNano()

	// LogQL query to search for logs from our service
	query := `{service_name="` + appName + `"}`

	// Query Loki using range query API
	url := fmt.Sprintf("http://localhost:3100/loki/api/v1/query_range?query=%s&start=%d&end=%d",
		query, start, end)

	resp, err := facades.Http().Get(url)
	s.NoError(err)

	body, err := resp.Body()
	s.NoError(err)

	// Verify logs contain expected content
	s.Contains(body, "test telemetry log")
}
