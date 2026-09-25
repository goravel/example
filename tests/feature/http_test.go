package feature

import (
	"fmt"
	"io"
	nethttp "net/http"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractstestinghttp "github.com/goravel/framework/contracts/testing/http"
	"github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

type HttpTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestHttpTestSuite(t *testing.T) {
	suite.Run(t, &HttpTestSuite{})
}

func (s *HttpTestSuite) SetupSuite() {
}

// SetupTest will run before each test in the suite.
func (s *HttpTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *HttpTestSuite) TearDownTest() {
}

func (s *HttpTestSuite) TestBindQuery() {
	resp, err := s.Http(s.T()).Get("/bind-query?name=Goravel")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"name\":\"Goravel\"}", content)
}

// bodyLimitByte is the app's configured request body limit and the drivers'
// fallback for body_limit <= 0: config/http.go sets body_limit to 4096 KiB.
const bodyLimitByte = 4096 << 10 // 4 MiB

// TestBodyLimit verifies http.drivers.<driver>.body_limit rejects an over-limit
// request with 413 before any handler runs (goravel/gin#247). The limit is left
// at the app default so the real 4 MiB boundary is exercised.
func (s *HttpTestSuite) TestBodyLimit() {
	s.Run("under the limit", func() {
		body, err := http.NewBody().SetField("test", map[string]any{"key": "value"}).Build()
		s.Require().NoError(err)

		resp, err := s.Http(s.T()).Post("/input-map", body.Reader())
		s.Require().NoError(err)
		resp.AssertSuccessful()

		content, err := resp.Content()
		s.Require().NoError(err)
		s.Equal(`{"test":{"key":"value"}}`, content)
	})

	s.Run("at the limit", func() {
		// A valid JSON body of exactly bodyLimitByte bytes is allowed: the
		// limit rejects only bodies strictly larger than it. The wrapper
		// `{"test":{"key":"..."}}` adds 19 bytes around the padding.
		atLimit := `{"test":{"key":"` + strings.Repeat("a", bodyLimitByte-19) + `"}}`
		s.Require().Len(atLimit, bodyLimitByte)

		resp, err := s.Http(s.T()).Post("/input-map", strings.NewReader(atLimit))
		s.Require().NoError(err)
		resp.AssertSuccessful()
	})

	over := `{"test":"` + strings.Repeat("a", bodyLimitByte+1) + `"}`
	formBody, err := http.NewBody(http.BodyTypeForm).
		SetField("test", strings.Repeat("a", bodyLimitByte+1)).
		Build()
	s.Require().NoError(err)
	multipartBody, err := http.NewBody(http.BodyTypeMultipart).
		SetField("test", strings.Repeat("a", bodyLimitByte+1)).
		Build()
	s.Require().NoError(err)

	cases := []struct {
		name        string
		path        string
		body        io.Reader
		contentType string
	}{
		{name: "json", path: "/input-map", body: strings.NewReader(over)},
		{name: "json to an unmatched route", path: "/not-a-route", body: strings.NewReader(over)},
		{name: "form", path: "/input-map", body: formBody.Reader(), contentType: formBody.ContentType()},
		{name: "multipart", path: "/input-map", body: multipartBody.Reader(), contentType: multipartBody.ContentType()},
	}

	for _, test := range cases {
		s.Run("over the limit "+test.name, func() {
			request := s.Http(s.T())
			if test.contentType != "" {
				request = request.WithHeader("Content-Type", test.contentType)
			}

			resp, err := request.Post(test.path, test.body)
			s.assertBodyRejected(resp, err)
		})
	}

	s.Run("over the limit chunked", func() {
		// fiber's in-process harness appends "Content-Length: -1" for a body with
		// no declared length, which fasthttp rejects before the limit applies, so
		// this case runs under gin only.
		switch driver := facades.Config().GetString("http.default"); driver {
		case "gin":
			// gin's in-process harness can express a chunked request.
		case "fiber":
			s.T().Skip("fiber's in-process harness cannot express a chunked request")
		default:
			s.T().Fatalf("unsupported http driver %q", driver)
		}

		// io.NopCloser hides the concrete reader from httptest.NewRequest, so the
		// request is sent without a Content-Length, like a chunked body.
		resp, err := s.Http(s.T()).Post("/input-map", io.NopCloser(strings.NewReader(over)))
		s.assertBodyRejected(resp, err)
	})
}

// TestBodyLimitDefaultFallback verifies body_limit 0 falls back to 4096 KiB
// on both drivers instead of disabling the limit.
func (s *HttpTestSuite) TestBodyLimitDefaultFallback() {
	scope, err := tests.OverrideConfig(map[string]any{
		"http.drivers.gin.body_limit":   0,
		"http.drivers.fiber.body_limit": 0,
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	// On fiber, body_limit 0 is passed straight through as fiber's BodyLimit,
	// and fiber treats 0 as its own 4 MiB default. The fiber half below
	// therefore verifies the observable contract (0 keeps the 4 MiB limit)
	// rather than goravel/fiber's <= 0 fallback mapping; the gin half verifies
	// the driver's fallback code.

	// 0 must not reject normal requests.
	body, err := http.NewBody().SetField("test", map[string]any{"key": "value"}).Build()
	s.Require().NoError(err)
	resp, err := s.Http(s.T()).Post("/input-map", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	// One byte over the 4 MiB fallback is rejected.
	over := `{"test":"` + strings.Repeat("a", bodyLimitByte+1) + `"}`
	resp, err = s.Http(s.T()).Post("/input-map", strings.NewReader(over))
	s.assertBodyRejected(resp, err)
}

// assertBodyRejected asserts that the driver rejected an over-limit request
// before any handler ran. gin answers with a 413 response. fiber's in-process
// harness (fiber.App.Test) rejects at the connection layer instead, so it
// returns no response and fasthttp.ErrBodyTooLarge. Any other driver fails the
// test rather than silently taking one of those branches.
func (s *HttpTestSuite) assertBodyRejected(resp contractstestinghttp.Response, err error) {
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
		s.T().Fatalf("unsupported http driver %q", driver)
	}
}

func (s *HttpTestSuite) TestFallback() {
	resp, err := s.Http(s.T()).Get("/lang")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).Get("/not-found")
	s.Require().NoError(err)
	resp.AssertNotFound()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("fallback", content)
}

func (s *HttpTestSuite) TestFiles() {
	body, err := http.NewBody().SetFiles(map[string][]string{
		"files": {"log_test.go", "support_test.go"},
	}).Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Post("/files", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"files\":[\"log_test.go\",\"support_test.go\"]}", content)
}

func (s *HttpTestSuite) TestInputMap() {
	body, err := http.NewBody().SetField("test", map[string]any{"key1": "value1", "key2": "value2"}).Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Post("/input-map", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"test\":{\"key1\":\"value1\",\"key2\":\"value2\"}}", content)
}

func (s *HttpTestSuite) TestInputMapArray() {
	body, err := http.NewBody().SetField("test", []map[string]any{{"key1": "value1", "key2": "value2"}, {"key3": "value3", "key4": "value4"}}).Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Post("/input-map-array", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"test\":[{\"key1\":\"value1\",\"key2\":\"value2\"},{\"key3\":\"value3\",\"key4\":\"value4\"}]}", content)
}

func (s *HttpTestSuite) TestLang() {
	tests := []struct {
		name           string
		lang           string
		expectResponse map[string]any
	}{
		{
			name:           "use default lang",
			expectResponse: map[string]any{"current_locale": "en", "fallback": "Goravel 是一个基于 Go 语言的 Web 开发框架", "name": "Goravel Framework"},
		},
		{
			name:           "lang is cn",
			lang:           "cn",
			expectResponse: map[string]any{"current_locale": "cn", "fallback": "Goravel 是一个基于 Go 语言的 Web 开发框架", "name": "Goravel 框架"},
		},
		{
			name:           "lang is fs",
			lang:           "fs",
			expectResponse: map[string]any{"current_locale": "fs", "fallback": "Goravel 是一个基于 Go 语言的 Web 开发框架", "name": "fs name"},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			resp, err := s.Http(s.T()).Get(fmt.Sprintf("/lang?lang=%s", test.lang))

			s.NoError(err)
			resp.AssertSuccessful()
			resp.AssertJson(test.expectResponse)
		})
	}
}

func (s *HttpTestSuite) TestPanic() {
	resp, err := s.Http(s.T()).Get("/panic")

	s.Require().NoError(err)
	resp.AssertInternalServerError()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("recover", content)
}

func (s *HttpTestSuite) TestStream() {
	resp, err := s.Http(s.T()).Get("/stream")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\n", content)
}

func (s *HttpTestSuite) TestThrottle() {
	// Clear cache to reset throttle count
	facades.Cache().Flush()

	resp, err := s.Http(s.T()).Get("/throttle")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).Get("/throttle")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).Get("/throttle")
	s.Require().NoError(err)
	resp.AssertTooManyRequests()
}

func (s *HttpTestSuite) TestTimeout() {
	resp, err := s.Http(s.T()).Get("/timeout")

	s.Require().NoError(err)
	resp.AssertStatus(contractshttp.StatusRequestTimeout)
}

func (s *HttpTestSuite) TestTimeoutIsolation() {
	timeoutResp, err := s.Http(s.T()).Get("/timeout-isolated?token=stale")

	s.Require().NoError(err)
	timeoutResp.AssertStatus(contractshttp.StatusRequestTimeout)

	timeoutContent, err := timeoutResp.Content()
	s.Require().NoError(err)
	s.Equal("Request Timeout", timeoutContent)

	freshResp, err := s.Http(s.T()).Get("/timeout-after?token=fresh")

	s.Require().NoError(err)
	freshResp.AssertSuccessful()

	freshContent, err := freshResp.Content()
	s.Require().NoError(err)
	s.Equal("{\"token\":\"fresh\"}", freshContent)
}

func (s *HttpTestSuite) TestUrl() {
	resp, err := s.Http(s.T()).Get("/url/get/1?a=1&b=2")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	var getResponse struct {
		FullURL string `json:"full_url"`
		Info    struct {
			Handler string `json:"handler"`
			Method  string `json:"method"`
			Name    string `json:"name"`
			Path    string `json:"path"`
		} `json:"info"`
		Info1 struct {
			Handler string `json:"handler"`
			Method  string `json:"method"`
			Name    string `json:"name"`
			Path    string `json:"path"`
		} `json:"info1"`
		Method     string `json:"method"`
		Name       string `json:"name"`
		OriginPath string `json:"origin_path"`
		Path       string `json:"path"`
		URL        string `json:"url"`
	}

	s.Require().NoError(resp.Bind(&getResponse))
	s.Equal("http://example.com/url/get/1?a=1&b=2", getResponse.FullURL)
	s.Equal("GET", getResponse.Info.Method)
	s.Equal("url.get", getResponse.Info.Name)
	s.Equal("/url/get/{id}", getResponse.Info.Path)
	s.Contains(getResponse.Info.Handler, "goravel/routes.Api.")
	s.Equal("GET|HEAD", getResponse.Info1.Method)
	s.Equal("url.get", getResponse.Info1.Name)
	s.Equal("/url/get/{id}", getResponse.Info1.Path)
	s.Contains(getResponse.Info1.Handler, "goravel/routes.Api.")
	s.Equal("GET", getResponse.Method)
	s.Equal("url.get", getResponse.Name)
	s.Equal("/url/get/{id}", getResponse.OriginPath)
	s.Equal("/url/get/1", getResponse.Path)
	s.Equal("/url/get/1?a=1&b=2", getResponse.URL)

	resp, err = s.Http(s.T()).Post("/url/post/1?a=1&b=2", strings.NewReader("{\"name\":\"Goravel\"}"))
	s.Require().NoError(err)
	resp.AssertSuccessful()

	var postResponse struct {
		FullURL string `json:"full_url"`
		Info    struct {
			Handler string `json:"handler"`
			Method  string `json:"method"`
			Name    string `json:"name"`
			Path    string `json:"path"`
		} `json:"info"`
		Info1 struct {
			Handler string `json:"handler"`
			Method  string `json:"method"`
			Name    string `json:"name"`
			Path    string `json:"path"`
		} `json:"info1"`
		Method     string `json:"method"`
		Name       string `json:"name"`
		OriginPath string `json:"origin_path"`
		Path       string `json:"path"`
		URL        string `json:"url"`
	}

	s.Require().NoError(resp.Bind(&postResponse))
	s.Equal("http://example.com/url/post/1?a=1&b=2", postResponse.FullURL)
	s.Equal("POST", postResponse.Info.Method)
	s.Equal("url.post", postResponse.Info.Name)
	s.Equal("/url/post/{id}", postResponse.Info.Path)
	s.Contains(postResponse.Info.Handler, "goravel/routes.Api.")
	s.Equal("POST", postResponse.Info1.Method)
	s.Equal("url.post", postResponse.Info1.Name)
	s.Equal("/url/post/{id}", postResponse.Info1.Path)
	s.Contains(postResponse.Info1.Handler, "goravel/routes.Api.")
	s.Equal("POST", postResponse.Method)
	s.Equal("url.post", postResponse.Name)
	s.Equal("/url/post/{id}", postResponse.OriginPath)
	s.Equal("/url/post/1", postResponse.Path)
	s.Equal("/url/post/1?a=1&b=2", postResponse.URL)
}

func (s *HttpTestSuite) TestUsers() {
	// Add a user
	var createdUser struct {
		User models.User
	}

	body, err := http.NewBody().SetField("name", "Goravel").SetField("avatar", "https://goravel.dev/avatar.png").Build()
	s.Require().NoError(err)
	resp, err := s.Http(s.T()).Post("users", body.Reader())

	s.Require().NoError(err)
	resp.AssertSuccessful()

	s.NoError(resp.Bind(&createdUser))
	s.True(createdUser.User.ID > 0)
	s.Equal("Goravel", createdUser.User.Name)
	s.Equal("https://goravel.dev/avatar.png", createdUser.User.Avatar)

	// Get Users
	var users struct {
		Users []models.User
	}
	resp, err = s.Http(s.T()).Get("users")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	s.NoError(resp.Bind(&users))
	s.Equal(1, len(users.Users))
	s.True(users.Users[0].ID > 0)
	s.Equal("Goravel", users.Users[0].Name)
	s.Equal("https://goravel.dev/avatar.png", users.Users[0].Avatar)

	// Update the User
	var updatedUser struct {
		User models.User
	}

	body, err = http.NewBody().SetField("name", "Framework").Build()
	s.Require().NoError(err)
	resp, err = s.Http(s.T()).Put(fmt.Sprintf("users/%d", createdUser.User.ID), body.Reader())

	s.Require().NoError(err)
	resp.AssertSuccessful()

	s.NoError(resp.Bind(&updatedUser))
	s.Equal(createdUser.User.ID, updatedUser.User.ID)
	s.Equal("Framework", updatedUser.User.Name)
	s.Equal("https://goravel.dev/avatar.png", updatedUser.User.Avatar)

	// Get the User
	var user struct {
		User models.User
	}
	resp, err = s.Http(s.T()).Get(fmt.Sprintf("users/%d", createdUser.User.ID))

	s.Require().NoError(err)
	resp.AssertSuccessful()

	s.NoError(resp.Bind(&user))
	s.True(user.User.ID > 0)
	s.Equal("Framework", user.User.Name)
	s.Equal("https://goravel.dev/avatar.png", user.User.Avatar)

	// Delete the User
	resp, err = s.Http(s.T()).Delete(fmt.Sprintf("users/%d", createdUser.User.ID), nil)

	s.Require().NoError(err)
	resp.AssertSuccessful()
	context, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"rows_affected\":1}", context)

	// Get Users
	resp, err = s.Http(s.T()).Get("users")

	s.Require().NoError(err)
	resp.AssertSuccessful()
	context, err = resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"users\":[]}", context)
}

func (s *HttpTestSuite) TestView() {
	resp, err := s.Http(s.T()).Get("/view")
	s.NoError(err)
	resp.AssertSuccessful()

	context, err := resp.Content()
	s.Require().NoError(err)

	csrfToken := resp.Headers().Get("X-CSRF-TOKEN")
	s.NotEmpty(csrfToken)
	s.Equal(context, fmt.Sprintf("\n  \n<html>\n  <body>\n    <p>I'm the header</p>\n\n  <p>Hello, Goravel</p>\n  <p> CSRF Token: %s </p>\n  \n  <p>I'm the footer</p>\n  </body>\n</html>\n\n", csrfToken))
}

func (s *HttpTestSuite) TestProviderRoute() {
	resp, err := s.Http(s.T()).Get("/provider-route")
	s.NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.NoError(err)
	s.Equal("Hello from provider route", content)
}
