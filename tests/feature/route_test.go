package feature

import (
	"fmt"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

	goraveltests "goravel/tests"
)

type RouteTestSuite struct {
	suite.Suite
	goraveltests.TestCase
}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, &RouteTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *RouteTestSuite) SetupTest() {
	//s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *RouteTestSuite) TearDownTest() {
}

//func (s *RouteTestSuite) TestAuth() {
//	type Response struct {
//		User models.User
//	}
//
//	tests := []struct {
//		name  string
//		guard string
//	}{
//		{
//			name: "default guard",
//		},
//		{
//			name:  "admin guard",
//			guard: "admin",
//		},
//	}
//
//	for _, test := range tests {
//		s.Run(test.name, func() {
//			// Unauthorized
//			resp, err := s.http.R().Get("auth/info")
//
//			s.Require().NoError(err)
//			s.Require().Equal(http.StatusUnauthorized, resp.StatusCode())
//
//			// Login
//			var authLogin Response
//			resp, err = s.http.R().SetResult(&authLogin).
//				SetHeader("Guard", test.guard).
//				SetBody(map[string]string{
//					"name": test.name,
//				}).Post("auth/login")
//
//			s.Require().NoError(err)
//			s.Require().Equal(http.StatusOK, resp.StatusCode())
//			s.True(authLogin.User.ID > 0)
//			s.Equal(test.name, authLogin.User.Name)
//
//			token := resp.Header().Get("Authorization")
//			s.Require().NotEmpty(token)
//
//			// Get User
//			var authUser Response
//			resp, err = s.http.R().SetResult(&authUser).SetHeaders(map[string]string{
//				"Authorization": token,
//				"Guard":         test.guard,
//			}).Get("auth/info")
//
//			s.Require().NoError(err)
//			s.Require().Equal(http.StatusOK, resp.StatusCode())
//			s.Equal(authLogin.User.ID, authUser.User.ID)
//			s.Equal(authLogin.User.Name, authUser.User.Name)
//		})
//	}
//}

func (s *RouteTestSuite) TestLang() {
	tests := []struct {
		name                  string
		lang                  string
		expectedCurrentLocale string
		expectedFallback      string
		expectedName          string
	}{
		{
			name:                  "use default lang",
			expectedCurrentLocale: "en",
			expectedFallback:      "Goravel 是一个基于 Go 语言的 Web 开发框架",
			expectedName:          "Goravel Framework",
		},
		{
			name:                  "lang is cn",
			lang:                  "cn",
			expectedCurrentLocale: "cn",
			expectedFallback:      "Goravel 是一个基于 Go 语言的 Web 开发框架",
			expectedName:          "Goravel 框架",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			response, err := s.Http(s.T()).Get(fmt.Sprintf("/lang?lang=%s", test.lang))

			s.NoError(err)
			response.AssertOk()
			response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
				json.Where("current_locale", test.expectedCurrentLocale).
					Where("name", test.expectedName).
					Where("fallback", test.expectedFallback)
			})
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
			var response contractstesting.TestResponse
			var err error
			for i := 0; i < 5; i++ {
				response, err = s.Http(s.T()).Get("/jwt/login")
				s.NoError(err)
			}
			response.AssertStatus(test.expectStatusCode)
		})
	}
}

func (s *RouteTestSuite) TestUsers() {
	// Add a user
	builder := http.NewBody().
		SetFields(map[string]any{
			"name":   "Goravel",
			"avatar": "https://goravel.dev/avatar.png",
		})

	body, err := builder.Build()
	s.NoError(err)

	response, err := s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Post("/users", body.Reader())

	s.NoError(err)
	response.AssertOk()
	//response.AssertJson(map[string]any{
	//	"user": map[string]any{
	//		"name":   "Goravel",
	//		"avatar": "https://goravel.dev/avatar.png",
	//	},
	//})

	json, err := response.Json()
	s.NoError(err)
	createdUser := json["user"].(map[string]any)

	s.True(createdUser["id"].(float64) > 0)
	s.Equal("Goravel", createdUser["Name"])
	s.Equal("https://goravel.dev/avatar.png", createdUser["Avatar"])

	// Get Users
	response, err = s.Http(s.T()).Get("/users")

	s.NoError(err)
	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Count("users", 1).
			First("users", func(user contractstesting.AssertableJSON) {
				user.Where("Name", "Goravel").
					Where("Avatar", "https://goravel.dev/avatar.png").
					WhereNot("id", 0)
			})
	})

	// Update the User
	builder = http.NewBody().
		SetFields(map[string]any{
			"name": "Framework",
		})

	body, err = builder.Build()
	s.NoError(err)

	response, err = s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Put(
		fmt.Sprintf("/users/%d", int(createdUser["id"].(float64))),
		body.Reader(),
	)

	s.NoError(err)
	response.AssertOk()
	//response.AssertJson(map[string]any{
	//	"user": map[string]any{
	//		"id":     createdUser["id"],
	//		"name":   "Framework",
	//		"avatar": "https://goravel.dev/avatar.png",
	//	},
	//})

	// Get the User
	response, err = s.Http(s.T()).Get(fmt.Sprintf("/users/%d", int(createdUser["id"].(float64))))

	s.NoError(err)
	response.AssertOk()
	response.AssertJson(map[string]any{
		"user": map[string]any{
			"id":     createdUser["id"],
			"name":   "Framework",
			"avatar": "https://goravel.dev/avatar.png",
		},
	})

	// Delete the User
	response, err = s.Http(s.T()).Delete(fmt.Sprintf("/users/%d", int(createdUser["id"].(float64))), nil)

	s.NoError(err)
	response.AssertOk()
	response.AssertJson(map[string]any{
		"rows_affected": float64(1),
	})

	// Confirm User Deletion
	response, err = s.Http(s.T()).Get("/users")

	s.NoError(err)
	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("users", []any{})
	})
}

func (s *RouteTestSuite) TestValidationJson() {
	builder := http.NewBody().
		SetFields(map[string]any{
			"name": "goravel",
			"date": "2024-07-08 18:33:32",
		})

	body, err := builder.Build()
	s.NoError(err)

	response, err := s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Post("/validation/json", body.Reader())

	s.NoError(err)
	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("name", "goravel").
			Where("date", "2024-07-08 18:33:32")
	})
}

func (s *RouteTestSuite) TestValidationRequest() {
	builder := http.NewBody().
		SetFields(map[string]any{
			"name":   "goravel",
			"date":   "2024-07-08 18:33:32",
			"tags":   []string{"tag1", "tag2"},
			"scores": []int{1, 2},
		})

	body, err := builder.Build()
	s.NoError(err)

	response, err := s.Http(s.T()).WithHeader("Content-Type", body.ContentType()).Post("/validation/request", body.Reader())

	s.NoError(err)
	response.AssertOk()
	response.AssertFluentJson(func(json contractstesting.AssertableJSON) {
		json.Where("name", "goravel").
			Where("date", "2024-07-08 18:33:32").
			Where("scores", []any{float64(1), float64(2)}).
			Where("tags", []any{"tag1", "tag2"})
	})
}
