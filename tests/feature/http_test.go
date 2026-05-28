package feature

import (
	"fmt"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

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
