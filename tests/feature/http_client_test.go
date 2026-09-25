package feature

import (
	"io"
	nethttp "net/http"
	"strings"
	"testing"

	client "github.com/goravel/framework/contracts/http/client"
	supporthttp "github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// smallBodyLimitKB keeps the over/under payloads tiny while still going through
// the real socket, so the client can read the 413 without an upload/close race.
const (
	smallBodyLimitKB   = 1
	smallBodyLimitByte = smallBodyLimitKB << 10
)

type HttpClientTestSuite struct {
	suite.Suite
	tests.TestCase

	scope *tests.ConfigScope
}

func TestHttpClientTestSuite(t *testing.T) {
	suite.Run(t, &HttpClientTestSuite{})
}

func (s *HttpClientTestSuite) SetupSuite() {
	// Lower body_limit for the whole suite so the real-client over-limit cases
	// send tiny bodies. OverrideConfig restarts the app, which also restarts the
	// framework HTTP runner serving 127.0.0.1:8080.
	scope, err := tests.OverrideConfig(map[string]any{
		"http.drivers.gin.body_limit":   smallBodyLimitKB,
		"http.drivers.fiber.body_limit": smallBodyLimitKB,
	})
	s.scope = scope
	if err != nil {
		// Restore before failing: testify skips TearDownSuite when SetupSuite
		// fails, and an unrestored 1 KiB limit would leak into later suites.
		s.NoError(scope.Restore())
		s.Require().FailNow("failed to set body_limit: " + err.Error())
	}
}

func (s *HttpClientTestSuite) TearDownSuite() {
	s.NoError(s.scope.Restore())
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

// TestBodyLimit verifies http.drivers.<driver>.body_limit rejects an over-limit
// request with 413 before any handler runs (goravel/gin#247). It drives the
// app's real HTTP client against the live server on 127.0.0.1:8080, so both
// drivers answer over a real socket.
func (s *HttpClientTestSuite) TestBodyLimit() {
	const jsonContentType = "application/json"

	s.Run("under the limit", func() {
		resp, err := facades.Http().WithHeader("Content-Type", jsonContentType).
			Post("/input-map", strings.NewReader(`{"test":{"key":"value"}}`))
		s.Require().NoError(err)
		s.Equal(nethttp.StatusOK, resp.Status())

		content, err := resp.Body()
		s.Require().NoError(err)
		s.Equal(`{"test":{"key":"value"}}`, content)
	})

	over := `{"test":"` + strings.Repeat("a", smallBodyLimitByte+1) + `"}`
	formBody := "test=" + strings.Repeat("a", smallBodyLimitByte+1)
	multipartBody, err := supporthttp.NewBody(supporthttp.BodyTypeMultipart).
		SetField("test", strings.Repeat("a", smallBodyLimitByte+1)).
		Build()
	s.Require().NoError(err)
	multipartPayload, err := io.ReadAll(multipartBody.Reader())
	s.Require().NoError(err)

	s.Run("at the limit", func() {
		// Exactly smallBodyLimitByte bytes of valid JSON is accepted (strict >).
		pad := smallBodyLimitByte - len(`{"test":""}`)
		body := `{"test":"` + strings.Repeat("a", pad) + `"}`
		s.Require().Len(body, smallBodyLimitByte)

		resp, err := facades.Http().WithHeader("Content-Type", jsonContentType).
			Post("/input-map", strings.NewReader(body))
		s.Require().NoError(err)
		s.Equal(nethttp.StatusOK, resp.Status())
	})

	cases := []struct {
		name        string
		path        string
		body        string
		contentType string
		chunked     bool
	}{
		{name: "json", path: "/input-map", body: over, contentType: jsonContentType},
		{name: "json to an unmatched route", path: "/not-a-route", body: over, contentType: jsonContentType},
		{name: "form", path: "/input-map", body: formBody, contentType: "application/x-www-form-urlencoded"},
		{name: "multipart", path: "/input-map", body: string(multipartPayload), contentType: multipartBody.ContentType()},
		{name: "chunked", path: "/input-map", body: over, contentType: jsonContentType, chunked: true},
	}

	for _, test := range cases {
		s.Run("over the limit "+test.name, func() {
			var body io.Reader = strings.NewReader(test.body)
			if test.chunked {
				// io.NopCloser hides the concrete reader, so net/http sends
				// Transfer-Encoding: chunked with no Content-Length.
				body = io.NopCloser(strings.NewReader(test.body))
			}

			resp, err := facades.Http().WithHeader("Content-Type", test.contentType).
				Post(test.path, body)
			s.assertBodyTooLarge(resp, err)
		})
	}
}

// assertBodyTooLarge asserts the real wire response for an over-limit request.
func (s *HttpClientTestSuite) assertBodyTooLarge(resp client.Response, err error) {
	s.T().Helper()

	s.Require().NoError(err)
	s.Equal(nethttp.StatusRequestEntityTooLarge, resp.Status())
	s.Equal("text/plain; charset=utf-8", resp.Headers().Get("Content-Type"))

	content, err := resp.Body()
	s.Require().NoError(err)
	s.Equal("Request Entity Too Large", content)
}

// TestBodyLimitDefaultFallback verifies body_limit 0 does not reject normal
// requests on either driver (gin maps <= 0 to the 4096 KiB default; fiber
// applies its own 4 MiB default). The 4 MiB boundary itself is not asserted
// here to avoid racing the server's Content-Length rejection during upload.
func (s *HttpClientTestSuite) TestBodyLimitDefaultFallback() {
	scope, err := tests.OverrideConfig(map[string]any{
		"http.drivers.gin.body_limit":   0,
		"http.drivers.fiber.body_limit": 0,
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	// 0 must not reject normal requests.
	resp, err := facades.Http().WithHeader("Content-Type", "application/json").
		Post("/input-map", strings.NewReader(`{"test":{"key":"value"}}`))
	s.Require().NoError(err)
	s.Equal(nethttp.StatusOK, resp.Status())
}
