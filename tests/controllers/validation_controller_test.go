package controllers

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

/*
*****************************************
 1. Please init .env file before running the test;
 2. Running the HTTP server in the mail_test.go file;
 3. An HTTP package(eg: net/http) is required for now, will optimize the test experience in this issue:
    https://github.com/goravel/goravel/issues/441

*****************************************
*/
type ValidationControllerTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestValidationControllerTestSuite(t *testing.T) {
	suite.Run(t, &ValidationControllerTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *ValidationControllerTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *ValidationControllerTestSuite) TearDownTest() {
}

func (s *ValidationControllerTestSuite) TestJson() {
	payload := strings.NewReader(`{
		"name": "Goravel",
		"date": "2024-07-08 18:33:32"
	}`)
	resp, err := http.Post(route("/validation/json"), "application/json", payload)
	s.Require().NoError(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal(string(body), "{\"date\":\"2024-07-08 18:33:32\",\"name\":\"Goravel\"}")
}

func (s *ValidationControllerTestSuite) TestRequest() {
	payload := strings.NewReader(`{
		"name": "Goravel",
		"date": "2024-07-08 18:33:32",
		"tags": ["tag1", "tag2"],
		"scores": [1, 2]
	}`)
	resp, err := http.Post(route("/validation/request"), "application/json", payload)
	s.Require().NoError(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal(string(body), "{\"date\":\"2024-07-08 18:33:32\",\"name\":\"Goravel\",\"scores\":[1,2],\"tags\":[\"tag1\",\"tag2\"]}")
}
