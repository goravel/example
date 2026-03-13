package feature

import (
	"strconv"
	"testing"

	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
	"goravel/tests"
)

type AuthTestSuite struct {
	suite.Suite
	tests.TestCase
}

type authLoginResponse struct {
	ID   uint        `json:"id"`
	User models.User `json:"user"`
}

type authPayload struct {
	Guard string `json:"guard"`
	Key   string `json:"key"`
}

type jwtStatusResponse struct {
	Check      bool        `json:"check"`
	Guest      bool        `json:"guest"`
	ID         uint        `json:"id"`
	User       models.User `json:"user"`
	Payload    authPayload `json:"payload"`
	ParseError string      `json:"parse_error"`
	RefreshErr string      `json:"refresh_error"`
	Error      string      `json:"error"`
}

type sessionUnsupportedResponse struct {
	ParseError   string `json:"parse_error"`
	RefreshError string `json:"refresh_error"`
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, &AuthTestSuite{})
}

func (s *AuthTestSuite) SetupTest() {
	s.RefreshDatabase()
	carbon.ClearTestNow()
}

func (s *AuthTestSuite) TearDownTest() {
	carbon.ClearTestNow()
}

func (s *AuthTestSuite) TestJwtStatusAndLogout() {
	tests := []struct {
		name          string
		guard         string
		expectedGuard string
	}{
		{
			name:          "default guard",
			expectedGuard: "user",
		},
		{
			name:          "admin guard",
			guard:         "admin",
			expectedGuard: "admin",
		},
		{
			name:          "agent guard",
			guard:         "agent",
			expectedGuard: "agent",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Step 1: Verify the user starts as unauthenticated for the selected guard.
			var initialStatus jwtStatusResponse
			resp, err := s.Http(s.T()).WithHeader("Guard", test.guard).Get("jwt/status")
			s.Require().NoError(err)
			resp.AssertSuccessful()
			s.NoError(resp.Bind(&initialStatus))
			s.False(initialStatus.Check)
			s.True(initialStatus.Guest)

			// Step 2: Confirm protected user info is denied before login.
			resp, err = s.Http(s.T()).WithHeader("Guard", test.guard).Get("jwt/info")
			s.Require().NoError(err)
			resp.AssertUnauthorized()

			// Step 3: Log in and capture the issued JWT.
			body, err := http.NewBody().SetField("name", test.name).Build()
			s.Require().NoError(err)

			var login authLoginResponse
			resp, err = s.Http(s.T()).WithHeader("Guard", test.guard).Post("jwt/login", body.Reader())
			s.Require().NoError(err)
			resp.AssertSuccessful()
			s.NoError(resp.Bind(&login))
			s.NotZero(login.User.ID)
			s.Equal(test.name, login.User.Name)

			token := resp.Headers().Get("Authorization")
			s.Require().NotEmpty(token)

			// Step 4: Verify authenticated status reflects the logged-in user and guard.
			var status jwtStatusResponse
			resp, err = s.Http(s.T()).WithHeader("Authorization", token).WithHeader("Guard", test.guard).Get("jwt/status")
			s.Require().NoError(err)
			resp.AssertSuccessful()
			s.NoError(resp.Bind(&status))
			s.True(status.Check)
			s.False(status.Guest)
			s.Equal(login.User.ID, status.ID)
			s.Equal(login.User.ID, status.User.ID)
			s.Equal(login.User.Name, status.User.Name)
			s.Equal(test.expectedGuard, status.Payload.Guard)
			s.Equal(login.User.ID, uintFromString(status.Payload.Key))

			// Step 5: Confirm authenticated user info is available after login.
			var info authLoginResponse
			resp, err = s.Http(s.T()).WithHeader("Authorization", token).WithHeader("Guard", test.guard).Get("jwt/info")
			s.Require().NoError(err)
			resp.AssertSuccessful()
			s.NoError(resp.Bind(&info))
			s.Equal(login.User.ID, info.ID)
			s.Equal(login.User.Name, info.User.Name)

			// Step 6: Log out and invalidate the current token.
			resp, err = s.Http(s.T()).WithHeader("Authorization", token).WithHeader("Guard", test.guard).Post("jwt/logout", nil)
			s.Require().NoError(err)
			resp.AssertSuccessful()

			// Step 7: Verify the disabled token can no longer access protected endpoints.
			var loggedOutStatus jwtStatusResponse
			resp, err = s.Http(s.T()).WithHeader("Authorization", token).WithHeader("Guard", test.guard).Get("jwt/status")
			s.Require().NoError(err)
			resp.AssertSuccessful()
			s.NoError(resp.Bind(&loggedOutStatus))
			s.Equal(frameworkerrors.AuthTokenDisabled.Error(), loggedOutStatus.ParseError)

			resp, err = s.Http(s.T()).WithHeader("Authorization", token).WithHeader("Guard", test.guard).Get("jwt/info")
			s.Require().NoError(err)
			resp.AssertUnauthorized()
		})
	}
}

func (s *AuthTestSuite) TestJwtRefresh() {
	// Step 1: Freeze time and log in to obtain the initial JWT.
	now := carbon.Now()
	carbon.SetTestNow(now)

	body, err := http.NewBody().SetField("name", "jwt-refresh").Build()
	s.Require().NoError(err)

	var login authLoginResponse
	resp, err := s.Http(s.T()).Post("jwt/login", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&login))

	token := resp.Headers().Get("Authorization")
	s.Require().NotEmpty(token)

	// Step 2: Refresh the token within the valid refresh window.
	carbon.SetTestNow(now.Copy().AddMinute())

	var refreshed jwtStatusResponse
	resp, err = s.Http(s.T()).WithHeader("Authorization", token).Post("jwt/refresh", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&refreshed))
	s.Equal("user", refreshed.Payload.Guard)
	s.Equal(login.User.ID, uintFromString(refreshed.Payload.Key))

	refreshedToken := resp.Headers().Get("Authorization")
	s.Require().NotEmpty(refreshedToken)
	s.NotEqual(token, refreshedToken)

	// Step 3: Verify middleware transparently renews an expired access token.
	carbon.SetTestNow(now.Copy().AddMinutes(61))

	var info authLoginResponse
	resp, err = s.Http(s.T()).WithHeader("Authorization", token).Get("jwt/info")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&info))
	s.Equal(login.User.ID, info.ID)

	middlewareRefreshedToken := resp.Headers().Get("Authorization")
	s.Require().NotEmpty(middlewareRefreshedToken)
	s.NotEqual(token, middlewareRefreshedToken)

	// Step 4: Confirm refresh fails once the refresh deadline has passed.
	carbon.SetTestNow(now.Copy().AddMinutes(20221))

	var refreshExpired jwtStatusResponse
	resp, err = s.Http(s.T()).WithHeader("Authorization", token).Post("jwt/refresh", nil)
	s.Require().NoError(err)
	resp.AssertUnauthorized()
	s.NoError(resp.Bind(&refreshExpired))
	s.Equal(frameworkerrors.AuthTokenExpired.Error(), refreshExpired.ParseError)
	s.Equal(frameworkerrors.AuthRefreshTimeExceeded.Error(), refreshExpired.RefreshErr)
}

func (s *AuthTestSuite) TestSessionStatusLoginUsingIDLogoutAndUnsupportedMethods() {
	// Step 1: Verify the session guard starts unauthenticated.
	var status jwtStatusResponse
	resp, err := s.Http(s.T()).Get("session/status")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&status))
	s.False(status.Check)
	s.True(status.Guest)

	// Step 2: Confirm protected session info is unavailable before login.
	resp, err = s.Http(s.T()).Get("session/info")
	s.Require().NoError(err)
	resp.AssertUnauthorized()

	// Step 3: Check unsupported session driver methods report the expected errors.
	var unsupported sessionUnsupportedResponse
	resp, err = s.Http(s.T()).Get("session/unsupported")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&unsupported))
	s.Equal(frameworkerrors.AuthUnsupportedDriverMethod.Args("session").Error(), unsupported.ParseError)
	s.Equal(frameworkerrors.AuthUnsupportedDriverMethod.Args("session").Error(), unsupported.RefreshError)

	// Step 4: Log in with the session guard and capture the returned cookies.
	body, err := http.NewBody().SetField("name", "session-user").Build()
	s.Require().NoError(err)

	var login authLoginResponse
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").Post("session/login", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&login))

	cookies := resp.Cookies()
	s.Require().NotEmpty(cookies)

	// Step 5: Verify the authenticated session status and user info.
	var sessionStatus jwtStatusResponse
	resp, err = s.Http(s.T()).WithCookies(cookies).Get("session/status")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&sessionStatus))
	s.True(sessionStatus.Check)
	s.False(sessionStatus.Guest)
	s.Equal(login.User.ID, sessionStatus.ID)
	s.Equal(login.User.Name, sessionStatus.User.Name)

	var info authLoginResponse
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/info")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&info))
	s.Equal(login.User.ID, info.ID)
	s.Equal(login.User.Name, info.User.Name)

	// Step 6: Log out and confirm the original session is cleared.
	resp, err = s.Http(s.T()).WithCookies(cookies).Post("session/logout", nil)
	s.Require().NoError(err)
	resp.AssertSuccessful()

	resp, err = s.Http(s.T()).WithCookies(cookies).Get("session/status")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&status))
	s.False(status.Check)
	s.True(status.Guest)

	// Step 7: Verify protected session info is denied after logout.
	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(cookies).Get("session/info")
	s.Require().NoError(err)
	resp.AssertUnauthorized()

	// Step 8: Log in by user ID and verify the restored authenticated session.
	body, err = http.NewBody().SetField("id", login.User.ID).Build()
	s.Require().NoError(err)

	var loginByID authLoginResponse
	resp, err = s.Http(s.T()).Post("session/login/id", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&loginByID))
	s.Equal(login.User.ID, loginByID.ID)
	s.Equal(login.User.Name, loginByID.User.Name)

	loginByIDCookies := resp.Cookies()
	s.Require().NotEmpty(loginByIDCookies)

	resp, err = s.Http(s.T()).WithCookies(loginByIDCookies).Get("session/status")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&sessionStatus))
	s.True(sessionStatus.Check)
	s.False(sessionStatus.Guest)
	s.Equal(login.User.ID, sessionStatus.ID)
	s.Equal(login.User.Name, sessionStatus.User.Name)

	resp, err = s.Http(s.T()).WithHeader("Guard", "session").WithCookies(loginByIDCookies).Get("session/info")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	s.NoError(resp.Bind(&info))
	s.Equal(login.User.ID, info.ID)
	s.Equal(login.User.Name, info.User.Name)
}

func uintFromString(value string) uint {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}

	return uint(parsed)
}
