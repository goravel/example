package feature

import (
	"os"
	"strings"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type HttpTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestHttpTestSuite(t *testing.T) {
	suite.Run(t, new(HttpTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *HttpTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *HttpTestSuite) TearDownTest() {
	s.NoError(os.RemoveAll("storage"))
}

func (s *HttpTestSuite) TestIndex() {
	response, err := s.Http(s.T()).Get("/test/index")
	s.NoError(err)

	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("message", "Welcome to the Index route")
	})
}

func (s *HttpTestSuite) TestShow() {
	response, err := s.Http(s.T()).Get("/test/show/1")
	s.NoError(err)

	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("id", "1").
			Where("message", "Item retrieved successfully")
	})
}

func (s *HttpTestSuite) TestCreate() {
	builder := http.NewBody().SetField("name", "Test Item")
	body, err := builder.Build()
	s.NoError(err)

	response, err := s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Post("/test/create", body.Reader())
	s.NoError(err)

	response.AssertCreated()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("message", "Item created")
		json.Where("data", map[string]any{"name": "Test Item"})
	})
}

func (s *HttpTestSuite) TestUpdate() {
	builder := http.NewBody().SetField("name", "Updated Item")
	body, err := builder.Build()
	s.NoError(err)

	response, err := s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Put("/test/update/1", body.Reader())
	s.NoError(err)

	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("message", "Item updated")
		json.Where("id", "1")
		json.Where("data", map[string]any{"name": "Updated Item"})
	})
}

func (s *HttpTestSuite) TestUpdateInvalidRequest() {
	response, err := s.Http(s.T()).Put("/test/update/1", strings.NewReader("{\"name\": \"goravel\""))
	s.NoError(err)

	response.AssertBadRequest()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("error", "Invalid request")
	})
}

func (s *HttpTestSuite) TestDelete() {
	response, err := s.Http(s.T()).Delete("/test/delete/1", nil)
	s.NoError(err)

	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("message", "Item deleted")
		json.Where("id", "1")
	})
}

func (s *HttpTestSuite) TestDeleteMissingID() {
	response, err := s.Http(s.T()).Delete("/test/delete", nil)
	s.NoError(err)

	response.AssertNotFound()
}

func (s *HttpTestSuite) TestCustomHeader() {
	response, err := s.Http(s.T()).Get("/test/custom-header")
	s.NoError(err)

	response.AssertOk()
	response.AssertHeader("X-Custom-Header", "CustomHeaderValue")
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("message", "Response with custom header")
	})
}

func (s *HttpTestSuite) TestEmptyResponse() {
	response, err := s.Http(s.T()).Get("/test/empty-response")
	s.NoError(err)

	response.AssertNoContent()
}

func (s *HttpTestSuite) TestAuthorizationWithoutAuth() {
	response, err := s.Http(s.T()).Post("/test/authorization", nil)
	s.NoError(err)

	response.AssertUnauthorized()
}
