package feature

import (
	"fmt"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
	"goravel/tests"
)

type RouteTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, &RouteTestSuite{})
}

func (s *RouteTestSuite) SetupSuite() {
	fmt.Println("RouteTestSuite", facades.Config().GetString("cache.default"))
}

// SetupTest will run before each test in the suite.
func (s *RouteTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *RouteTestSuite) TearDownTest() {
}

func (s *RouteTestSuite) TestAuth() {
	type Response struct {
		ID   uint
		User models.User
	}

	tests := []struct {
		name  string
		guard string
	}{
		{
			name: "default guard",
		},
		{
			name:  "admin guard",
			guard: "admin",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Unauthorized
			resp, err := s.Http(s.T()).Get("auth/info")

			s.Require().NoError(err)
			resp.AssertUnauthorized()
			content, err := resp.Content()
			s.Require().NoError(err)
			s.Equal("Unauthorized", content)

			// Login
			var authLogin Response
			body, err := http.NewBody().SetField("name", test.name).Build()
			s.Require().NoError(err)
			resp, err = s.Http(s.T()).WithHeader("Guard", test.guard).Bind(&authLogin).Post("auth/login", body.Reader())

			s.Require().NoError(err)
			resp.AssertSuccessful()

			s.True(authLogin.User.ID > 0)
			s.Equal(test.name, authLogin.User.Name)

			token := resp.Headers().Get("Authorization")
			s.Require().NotEmpty(token)

			// Get User
			var authUser Response
			resp, err = s.Http(s.T()).WithHeader("Authorization", token).WithHeader("Guard", test.guard).Bind(&authUser).Get("auth/info")

			s.Require().NoError(err)
			resp.AssertSuccessful()
			s.Equal(authLogin.User.ID, authUser.User.ID)
			s.Equal(authLogin.User.Name, authUser.User.Name)
			s.Equal(authLogin.User.ID, authUser.ID)
		})
	}
}

func (s *RouteTestSuite) TestBindQuery() {
	resp, err := s.Http(s.T()).Get("/bind-query?name=Goravel")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"name\":\"Goravel\"}", content)
}

func (s *RouteTestSuite) TestLang() {
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

func (s *RouteTestSuite) TestPanic() {
	resp, err := s.Http(s.T()).Get("/panic")

	s.Require().NoError(err)
	resp.AssertInternalServerError()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Empty(content)
}

func (s *RouteTestSuite) TestStream() {
	resp, err := s.Http(s.T()).Get("/stream")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("a\nb\nc\n", content)
}

func (s *RouteTestSuite) TestThrottle() {
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

func (s *RouteTestSuite) TestTimeout() {
	resp, err := s.Http(s.T()).Get("/timeout")

	s.Require().NoError(err)
	resp.AssertStatus(contractshttp.StatusRequestTimeout)
}

func (s *RouteTestSuite) TestUsers() {
	// Add a user
	var createdUser struct {
		User models.User
	}

	body, err := http.NewBody().SetField("name", "Goravel").SetField("avatar", "https://goravel.dev/avatar.png").Build()
	s.Require().NoError(err)
	resp, err := s.Http(s.T()).Bind(&createdUser).Post("users", body.Reader())

	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.True(createdUser.User.ID > 0)
	s.Equal("Goravel", createdUser.User.Name)
	s.Equal("https://goravel.dev/avatar.png", createdUser.User.Avatar)

	// Get Users
	var users struct {
		Users []models.User
	}
	resp, err = s.Http(s.T()).Bind(&users).Get("users")

	s.Require().NoError(err)
	resp.AssertSuccessful()
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
	resp, err = s.Http(s.T()).Bind(&updatedUser).Put(fmt.Sprintf("users/%d", createdUser.User.ID), body.Reader())

	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.Equal(createdUser.User.ID, updatedUser.User.ID)
	s.Equal("Framework", updatedUser.User.Name)
	s.Equal("https://goravel.dev/avatar.png", updatedUser.User.Avatar)

	// Get the User
	var user struct {
		User models.User
	}
	resp, err = s.Http(s.T()).Bind(&user).Get(fmt.Sprintf("users/%d", createdUser.User.ID))

	s.Require().NoError(err)
	resp.AssertSuccessful()
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

func (s *RouteTestSuite) TestValidationJson() {
	payload := strings.NewReader(`{
		"name": "Goravel",
		"date": "2024-07-08 18:33:32"
	}`)

	resp, err := s.Http(s.T()).Post("/validation/json", payload)

	s.NoError(err)
	resp.AssertSuccessful()
	context, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("{\"date\":\"2024-07-08 18:33:32\",\"name\":\"Goravel\"}", context)
}

func (s *RouteTestSuite) TestValidationRequest() {
	s.Run("success", func() {
		payload := strings.NewReader(`{
			"name": " Goravel ",
			"date": "2024-07-08 18:33:32",
			"tags": ["tag1", "tag2"],
			"scores": [1, 2],
			"code": 123456
		}`)

		resp, err := s.Http(s.T()).Post("/validation/request", payload)

		s.NoError(err)
		resp.AssertSuccessful()
		context, err := resp.Content()
		s.Require().NoError(err)
		s.Equal("{\"code\":123456,\"date\":\"2024-07-08 18:33:32\",\"name\":\"Goravel\",\"scores\":[1,2],\"tags\":[\"tag1\",\"tag2\"]}", context)
	})

	s.Run("failed", func() {
		payload := strings.NewReader(`{
			"date": "1",
			"tags": "tag1",
			"scores": 1,
			"code": 1234567
		}`)

		resp, err := s.Http(s.T()).Post("/validation/request", payload)

		s.NoError(err)
		resp.AssertBadRequest()

		content, err := resp.Content()
		s.Require().NoError(err)
		s.Equal("{\"message\":{\"code\":{\"regex\":\"code value does not pass the regex check\"},\"date\":{\"date\":\"date value should be a date string\"},\"name\":{\"required\":\"name is required to not be empty\"}}}", content)
	})
}
