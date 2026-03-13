package feature

import (
	nethttp "net/http"

	"github.com/goravel/framework/support/http"
	"github.com/spf13/cast"

	"goravel/app/models"
)

type authLoginResponse struct {
	Guard string
	Token string
	User  models.User
}

type authStateResponse struct {
	Check   bool
	Guard   string
	Guest   bool
	ID      string
	IDError bool `json:"id_error"`
}

type authParseResponse struct {
	Error bool
	Guard string
	Key   string
}

type authRefreshResponse struct {
	Error bool
	Token string
}

func (s *HttpTestSuite) loginByID(path, guard, name string) (authLoginResponse, []*nethttp.Cookie) {
	body, err := http.NewBody().SetField("name", name).Build()
	s.Require().NoError(err)

	req := s.Http(s.T())
	if guard != "" {
		req = req.WithHeader("Guard", guard)
	}

	resp, err := req.Post(path, body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	var login authLoginResponse
	s.NoError(resp.Bind(&login))

	return login, resp.Cookies()
}

func (s *HttpTestSuite) TestAuthByJwt() {
	s.Run("login using id", func() {
		login, _ := s.loginByID("jwt/login-id", "", "jwt-login-id")
		s.Equal("user", login.Guard)
		s.NotEmpty(login.Token)
		s.True(login.User.ID > 0)
	})

	s.Run("parse and state", func() {
		login, _ := s.loginByID("jwt/login-id", "", "jwt-parse-state")

		var parse authParseResponse
		resp, err := s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Post("jwt/parse", nil)
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&parse))
		s.False(parse.Error)
		s.Equal(login.Guard, parse.Guard)
		s.Equal(cast.ToString(login.User.ID), parse.Key)

		var state authStateResponse
		resp, err = s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Get("jwt/state")
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&state))
		s.True(state.Check)
		s.False(state.Guest)
		s.False(state.IDError)
		s.Equal(cast.ToString(login.User.ID), state.ID)
	})

	s.Run("user info", func() {
		login, _ := s.loginByID("jwt/login-id", "", "jwt-user-info")

		var info struct {
			ID   uint
			User models.User
		}
		resp, err := s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Get("jwt/info")
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&info))
		s.Equal(login.User.ID, info.ID)
		s.Equal(login.User.ID, info.User.ID)
	})

	s.Run("refresh and logout", func() {
		login, _ := s.loginByID("jwt/login-id", "", "jwt-refresh")

		var refresh authRefreshResponse
		resp, err := s.Http(s.T()).WithHeader("Authorization", "Bearer "+login.Token).Post("jwt/refresh", nil)
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
	})
}

func (s *HttpTestSuite) TestAuthBySession() {
	s.Run("state before login", func() {
		var state authStateResponse
		resp, err := s.Http(s.T()).WithHeader("Guard", "session").Get("session/state")
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&state))
		s.False(state.Check)
		s.True(state.Guest)
		s.True(state.IDError)
	})

	s.Run("login using id and state", func() {
		login, cookies := s.loginByID("session/login-id", "session", "session-login-id")
		s.Equal("session", login.Guard)
		s.Empty(login.Token)
		s.True(login.User.ID > 0)

		var state authStateResponse
		resp, err := s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/state")
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&state))
		s.True(state.Check)
		s.False(state.Guest)
		s.False(state.IDError)
		s.Equal(cast.ToString(login.User.ID), state.ID)
	})

	s.Run("user info and unsupported parse/refresh", func() {
		login, cookies := s.loginByID("session/login-id", "session", "session-user-info")

		var info struct {
			ID   uint
			User models.User
		}
		resp, err := s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/info")
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&info))
		s.Equal(login.User.ID, info.ID)
		s.Equal(login.User.ID, info.User.ID)

		var parse authRefreshResponse
		resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Post("session/parse", nil)
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&parse))
		s.True(parse.Error)

		var refresh authRefreshResponse
		resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Post("session/refresh", nil)
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&refresh))
		s.True(refresh.Error)
		s.Empty(refresh.Token)
	})

	s.Run("logout", func() {
		_, cookies := s.loginByID("session/login-id", "session", "session-logout")

		resp, err := s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Post("session/logout", nil)
		s.Require().NoError(err)
		resp.AssertSuccessful()

		var state authStateResponse
		resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/state")
		s.Require().NoError(err)
		resp.AssertSuccessful()
		s.NoError(resp.Bind(&state))
		s.False(state.Check)
		s.True(state.Guest)
		s.True(state.IDError)
	})
}
