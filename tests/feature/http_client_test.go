package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

type HttpClientTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestHttpClientTestSuite(t *testing.T) {
	suite.Run(t, &HttpClientTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *HttpClientTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *HttpClientTestSuite) TearDownTest() {
}

func (s *HttpClientTestSuite) TestGet() {
	response, err := facades.Http().Get("/users")
	s.Require().NoError(err)
	s.Equal(200, response.Status())
}
