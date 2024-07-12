package controllers

import (
	"github.com/goravel/framework/support/debug"
	"io"
	"net/http"
	"sync"
	"testing"

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
			_, err := http.Get(route("/session/put"))
			s.Require().NoError(err)

			var wg sync.WaitGroup
			for i := 0; i < 1000; i++ {
				wg.Add(1)
				go func() {
					resp, err := http.Get(route("/session/get"))
					s.Require().NoError(err)
					defer resp.Body.Close()
					body, err := io.ReadAll(resp.Body)

					s.Require().NoError(err)
					s.Equal(http.StatusOK, resp.StatusCode)
					s.Equal(test.expectResponse, string(body))

					if test.expectResponse != string(body) {
						debug.Dump(string(body))
					}

					wg.Done()
				}()
			}
		})
	}
}
