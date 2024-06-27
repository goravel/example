package middlewares

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type ThrottleTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestThrottleTestSuite(t *testing.T) {
	suite.Run(t, &ThrottleTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *ThrottleTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *ThrottleTestSuite) TearDownTest() {
}

func (s *ThrottleTestSuite) TestThrottle() {
	tests := []struct {
		name             string
		expectStatusCode int
	}{
		{
			name:             "no throttle",
			expectStatusCode: 200,
		},
		{
			name:             "throttle",
			expectStatusCode: 429,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			var resp *http.Response
			var err error
			for i := 0; i < 5; i++ {
				resp, err = http.Get(route("/jwt/login"))
				s.Require().NoError(err)
			}
			s.Equal(test.expectStatusCode, resp.StatusCode)
		})
	}
}
