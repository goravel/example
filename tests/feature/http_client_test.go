package feature

import (
	"io"
	nethttp "net/http"
	"strings"
	"testing"

	client "github.com/goravel/framework/contracts/http/client"
	contractshttptest "github.com/goravel/framework/contracts/testing/http"
	supporthttp "github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"goravel/app/facades"
	"goravel/tests"
)

// smallBodyLimitKB keeps the over/under payloads tiny while still going through
// the real socket, so the client can read the 413 without an upload/close race.
const (
	smallBodyLimitKB     = 1
	smallBodyLimitByte   = smallBodyLimitKB << 10
	defaultBodyLimitByte = 4096 << 10 // drivers' <= 0 fallback / config/http.go default
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

// TestBodyLimitDefaultFallback verifies body_limit 0 keeps the 4096 KiB limit
// on both drivers. Only gin maps <= 0 to 4096 KiB; goravel/fiber passes 0
// straight through and fiber itself defaults to 4 MiB, so the gin half
// verifies the fallback code while the fiber half verifies the observable
// contract. It stays on the in-process harness: a 4 MiB body over a real socket
// can be cut off by the server's early rejection, and the point here is the
// mapping, not the wire 413.
func (s *HttpClientTestSuite) TestBodyLimitDefaultFallback() {
	scope, err := tests.OverrideConfig(map[string]any{
		"http.drivers.gin.body_limit":   0,
		"http.drivers.fiber.body_limit": 0,
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	// 0 must not reject normal requests.
	body, err := supporthttp.NewBody().SetField("test", map[string]any{"key": "value"}).Build()
	s.Require().NoError(err)
	resp, err := s.Http(s.T()).Post("/input-map", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	// One byte over the 4 MiB fallback is rejected.
	over := `{"test":"` + strings.Repeat("a", defaultBodyLimitByte+1) + `"}`
	resp, err = s.Http(s.T()).Post("/input-map", strings.NewReader(over))
	s.assertBodyRejected(resp, err)
}

// assertBodyRejected covers the in-process harness, where gin answers with a
// 413 response and fiber returns fasthttp.ErrBodyTooLarge with no response.
func (s *HttpClientTestSuite) assertBodyRejected(resp contractshttptest.Response, err error) {
	s.T().Helper()

	switch driver := facades.Config().GetString("http.default"); driver {
	case "gin":
		s.Require().NoError(err)
		resp.AssertStatus(nethttp.StatusRequestEntityTooLarge)
		resp.AssertHeader("Content-Type", "text/plain; charset=utf-8")
		content, err := resp.Content()
		s.Require().NoError(err)
		s.Equal("Request Entity Too Large", content)
	case "fiber":
		s.Require().ErrorIs(err, fasthttp.ErrBodyTooLarge)
		s.Nil(resp)
	default:
		s.Require().FailNow("unsupported http driver: " + driver)
	}
}
