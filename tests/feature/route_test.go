package feature

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
	"goravel/tests"
)

type RouteTestSuite struct {
	suite.Suite
	tests.TestCase
	http *resty.Client
}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, &RouteTestSuite{
		http: resty.New().
			SetBaseURL(fmt.Sprintf("http://%s:%s",
				facades.Config().GetString("APP_HOST"),
				facades.Config().GetString("APP_PORT"))).
			SetHeader("Content-Type", "application/json"),
	})
}

// SetupTest will run before each test in the suite.
func (s *RouteTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *RouteTestSuite) TearDownTest() {
}

func (s *RouteTestSuite) TestLang() {
	tests := []struct {
		name           string
		lang           string
		expectResponse string
	}{
		{
			name:           "use default lang",
			expectResponse: "{\"current_locale\":\"en\",\"fallback\":\"Goravel 是一个基于 Go 语言的 Web 开发框架\",\"name\":\"Goravel Framework\"}",
		},
		{
			name:           "lang is cn",
			lang:           "cn",
			expectResponse: "{\"current_locale\":\"cn\",\"fallback\":\"Goravel 是一个基于 Go 语言的 Web 开发框架\",\"name\":\"Goravel 框架\"}",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			resp, err := s.http.R().Get(fmt.Sprintf("/lang?lang=%s", test.lang))

			s.NoError(err)
			s.Equal(http.StatusOK, resp.StatusCode())
			s.Equal(test.expectResponse, resp.String())
		})
	}
}

func (s *RouteTestSuite) TestThrottle() {
	tests := []struct {
		name             string
		expectStatusCode int
	}{
		{
			name:             "no throttle",
			expectStatusCode: 200,
		},
		{
			name:             "throttle",
			expectStatusCode: 429,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			var resp *resty.Response
			var err error
			for i := 0; i < 5; i++ {
				resp, err = s.http.R().Get("/jwt/login")
				s.Require().NoError(err)
			}
			s.Equal(test.expectStatusCode, resp.StatusCode())
		})
	}
}

func (s *RouteTestSuite) TestUsers() {
	// Add a user
	var createdUser struct {
		User models.User
	}

	resp, err := s.http.R().SetResult(&createdUser).SetBody(map[string]string{
		"name":   "Goravel",
		"avatar": "https://goravel.dev/avatar.png",
	}).Post("users")

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.True(createdUser.User.ID > 0)
	s.Equal("Goravel", createdUser.User.Name)
	s.Equal("https://goravel.dev/avatar.png", createdUser.User.Avatar)

	// Get Users
	var users struct {
		Users []models.User
	}
	resp, err = s.http.R().SetResult(&users).Get("users")

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.Equal(1, len(users.Users))
	s.True(createdUser.User.ID > 0)
	s.Equal("Goravel", users.Users[0].Name)
	s.Equal("https://goravel.dev/avatar.png", users.Users[0].Avatar)

	// Update the User
	var updatedUser struct {
		User models.User
	}

	resp, err = s.http.R().SetResult(&updatedUser).SetBody(map[string]string{
		"name": "Framework",
	}).Put(fmt.Sprintf("users/%d", createdUser.User.ID))

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.Equal(createdUser.User.ID, updatedUser.User.ID)
	s.Equal("Framework", updatedUser.User.Name)
	s.Equal("https://goravel.dev/avatar.png", updatedUser.User.Avatar)

	// Get the User
	var user struct {
		User models.User
	}
	resp, err = s.http.R().SetResult(&user).Get(fmt.Sprintf("users/%d", createdUser.User.ID))

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.True(user.User.ID > 0)
	s.Equal("Framework", user.User.Name)
	s.Equal("https://goravel.dev/avatar.png", user.User.Avatar)

	// Delete the User
	resp, err = s.http.R().Delete(fmt.Sprintf("users/%d", createdUser.User.ID))

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.Equal("{\"rows_affected\":1}", resp.String())

	// Get Users
	resp, err = s.http.R().Get("users")

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.Equal("{\"users\":[]}", resp.String())
}

func (s *RouteTestSuite) TestValidationJson() {
	payload := strings.NewReader(`{
		"name": "Goravel",
		"date": "2024-07-08 18:33:32"
	}`)

	resp, err := s.http.R().SetBody(payload).Post("/validation/json")

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
	s.Equal("{\"date\":\"2024-07-08 18:33:32\",\"name\":\"Goravel\"}", resp.String())
}

func (s *RouteTestSuite) TestValidationRequest() {
	payload := strings.NewReader(`{
		"name": "Goravel",
		"date": "2024-07-08 18:33:32",
		"tags": ["tag1", "tag2"],
		"scores": [1, 2]
	}`)

	resp, err := s.http.R().SetBody(payload).Post("/validation/request")

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
	s.Equal("{\"date\":\"2024-07-08 18:33:32\",\"name\":\"Goravel\",\"scores\":[1,2],\"tags\":[\"tag1\",\"tag2\"]}", resp.String())
}
