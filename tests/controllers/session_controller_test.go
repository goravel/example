package controllers

import (
	"github.com/goravel/framework/support/debug"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

/*
*****************************************
We need add the lang folder in the testing package for now, will optimize it in v1.15
*****************************************
*/
type SessionTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, &SessionTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *SessionTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *SessionTestSuite) TearDownTest() {
}

func (s *SessionTestSuite) TestIndex() {
	tests := []struct {
		name           string
		lang           string
		expectResponse string
	}{
		{
			name:           "has session",
			expectResponse: "{\"name\":\"Goravel\"}",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			resp, err := http.Get(route("/session/put"))
			s.Require().NoError(err)

			cookies := resp.Cookies()

			ticker := time.NewTicker(60 * time.Second)
			// 10min timeout
			after := time.After(600 * time.Second)

			for {
				select {
				case <-after:
					ticker.Stop()
					return
				case <-ticker.C:
					client := &http.Client{}
					var req *http.Request
					req, err = http.NewRequest("GET", route("/session/get"), nil)
					for _, v := range cookies {
						req.AddCookie(v)
					}

					resp, err := client.Do(req)
					s.Require().NoError(err)
					defer resp.Body.Close()
					body, err := io.ReadAll(resp.Body)
					cookies = resp.Cookies()

					s.Require().NoError(err)
					s.Equal(http.StatusOK, resp.StatusCode)
					s.Equal(test.expectResponse, string(body))

					if test.expectResponse != string(body) {
						debug.Dump(string(body))
					}
				}

			}
		})
	}
}
