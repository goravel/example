package feature

import (
	"fmt"
	"net/http"
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
}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, &RouteTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *RouteTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *RouteTestSuite) TearDownTest() {
}

func (s *RouteTestSuite) TestUsers() {
	client := resty.New().
		SetBaseURL(fmt.Sprintf("http://%s:%s",
			facades.Config().GetString("APP_HOST"),
			facades.Config().GetString("APP_PORT"))).
		SetHeader("Content-Type", "application/json")

	// Add a user
	var createdUser struct {
		User models.User
	}

	resp, err := client.R().SetResult(&createdUser).SetBody(map[string]string{
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
	resp, err = client.R().SetResult(&users).Get("users")

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

	resp, err = client.R().SetResult(&updatedUser).SetBody(map[string]string{
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
	resp, err = client.R().SetResult(&user).Get(fmt.Sprintf("users/%d", createdUser.User.ID))

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.True(user.User.ID > 0)
	s.Equal("Framework", user.User.Name)
	s.Equal("https://goravel.dev/avatar.png", user.User.Avatar)

	// Delete the User
	resp, err = client.R().Delete(fmt.Sprintf("users/%d", createdUser.User.ID))

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.Equal("{\"rows_affected\":1}", resp.String())

	// Get Users
	resp, err = client.R().Get("users")

	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode())
	s.Equal("{\"users\":[]}", resp.String())
}
