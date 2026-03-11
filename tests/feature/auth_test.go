package feature

import (
	"github.com/goravel/framework/support/http"
	"github.com/spf13/cast"

	"goravel/app/models"
)

func (s *HttpTestSuite) TestAuthContractsByJwt() {
	type LoginResponse struct {
		Guard string
		Token string
		User  models.User
	}
	type ParseResponse struct {
		Error bool
		Guard string
		Key   string
	}
	type StateResponse struct {
		Check   bool
		Guard   string
		Guest   bool
		ID      string
		IDError bool `json:"id_error"`
	}
	type RefreshResponse struct {
		Error bool
		Token string
	}

	body, err := http.NewBody().SetField("name", "jwt-login-id").Build()
	s.Require().NoError(err)

	var login LoginResponse
	resp, err := s.Http(s.T()).Post("jwt/login-id", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&login))
	s.Equal("user", login.Guard)
	s.NotEmpty(login.Token)
	s.True(login.User.ID > 0)

	var parse ParseResponse
	resp, err = s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Post("jwt/parse", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&parse))
	s.False(parse.Error)
	s.Equal(login.Guard, parse.Guard)
	s.Equal(cast.ToString(login.User.ID), parse.Key)

	var state StateResponse
	resp, err = s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Get("jwt/state")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&state))
	s.True(state.Check)
	s.False(state.Guest)
	s.False(state.IDError)
	s.Equal(cast.ToString(login.User.ID), state.ID)

	var info struct {
		ID   uint
		User models.User
	}
	resp, err = s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Get("jwt/info")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&info))
	s.Equal(login.User.ID, info.ID)
	s.Equal(login.User.ID, info.User.ID)

	var refresh RefreshResponse
	resp, err = s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Post("jwt/refresh", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&refresh))
	s.False(refresh.Error)
	s.NotEmpty(refresh.Token)

	resp, err = s.Http(s.T()).WithHeader("Authorization", "Bearer "+refresh.Token).Post("jwt/logout", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).WithHeader("Authorization", "Bearer "+refresh.Token).Get("jwt/info")
	s.Require().NoError(err)
	resp.AssertUnauthorized()
}

func (s *HttpTestSuite) TestAuthContractsBySession() {
	type LoginResponse struct {
		Guard string
		Token string
		User  models.User
	}
	type StateResponse struct {
		Check   bool
		Guard   string
		Guest   bool
		ID      string
		IDError bool `json:"id_error"`
	}
	type DriverResponse struct {
		Error bool
		Token string
	}

	var unauthenticatedState StateResponse
	resp, err := s.Http(s.T()).WithHeader("Guard", "session").Get("session/state")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&unauthenticatedState))
	s.False(unauthenticatedState.Check)
	s.True(unauthenticatedState.Guest)
	s.True(unauthenticatedState.IDError)

	body, err := http.NewBody().SetField("name", "session-login-id").Build()
	s.Require().NoError(err)

	var login LoginResponse
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").Post("session/login-id", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&login))
	s.Equal("session", login.Guard)
	s.Empty(login.Token)
	s.True(login.User.ID > 0)

	cookies := resp.Cookies()

	var authenticatedState StateResponse
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/state")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&authenticatedState))
	s.True(authenticatedState.Check)
	s.False(authenticatedState.Guest)
	s.False(authenticatedState.IDError)
	s.Equal(cast.ToString(login.User.ID), authenticatedState.ID)

	var info struct {
		ID   uint
		User models.User
	}
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/info")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&info))
	s.Equal(login.User.ID, info.ID)
	s.Equal(login.User.ID, info.User.ID)

	var parse DriverResponse
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Post("session/parse", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&parse))
	s.True(parse.Error)

	var refresh DriverResponse
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Post("session/refresh", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&refresh))
	s.True(refresh.Error)
	s.Empty(refresh.Token)

	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Post("session/logout", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()

	var loggedOutState StateResponse
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/state")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&loggedOutState))
	s.False(loggedOutState.Check)
	s.True(loggedOutState.Guest)
	s.True(loggedOutState.IDError)
}
