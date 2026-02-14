package feature

import (
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

func (s *TelemetryTestSuite) TestTelemetry() {
	appName := facades.Config().GetString("app.name")
	resp, err := s.Http(s.T()).Get("/users/1")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	time.Sleep(7 * time.Second)

	s.Run("Check Jaeger for Traces", func() {
		resp, err := facades.Http().Get("http://localhost:16686/api/traces?service=" + appName)
		s.NoError(err)

		var result struct {
			Data []any `json:"data"`
		}
		s.NoError(resp.Bind(&result))
		s.NotEmpty(result.Data, "Telemetry failed to reach Jaeger")
	})
}
