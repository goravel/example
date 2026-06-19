package feature

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

type MaintenanceTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestMaintenanceTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceTestSuite))
}

func (s *MaintenanceTestSuite) SetupTest() {
	_ = facades.Artisan().Call("up")
}

func (s *MaintenanceTestSuite) TearDownTest() {
	_ = facades.Artisan().Call("up")
}

func (s *MaintenanceTestSuite) TestNotInMaintenance_PassesThrough() {
	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *MaintenanceTestSuite) TestDownCommand_Basic() {
	err := facades.Artisan().Call("down")
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertStatus(http.StatusServiceUnavailable)

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("The application is under maintenance", content)
}

func (s *MaintenanceTestSuite) TestDownCommand_CustomReasonAndStatus() {
	err := facades.Artisan().Call("down --reason=Upgrading --status=500")
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertStatus(500)

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("Upgrading", content)
}

func (s *MaintenanceTestSuite) TestDownCommand_Redirect() {
	err := facades.Artisan().Call("down --redirect=/maintenance")
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertStatus(http.StatusTemporaryRedirect)

	s.Equal("/maintenance", resp.Headers().Get("Location"))
}

func (s *MaintenanceTestSuite) TestDownCommand_Render() {
	err := facades.Artisan().Call("down --render=welcome.tmpl")
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertStatus(http.StatusServiceUnavailable)

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Contains(content, "Goravel")
}

func (s *MaintenanceTestSuite) TestDownCommand_SecretBypass() {
	err := facades.Artisan().Call("down --secret=let-me-in")
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertStatus(http.StatusServiceUnavailable)

	resp, err = s.Http(s.T()).Get("/?secret=let-me-in")
	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *MaintenanceTestSuite) TestDownCommand_WithSecret_NotMatch() {
	err := facades.Artisan().Call("down --secret=let-me-in")
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Get("/?secret=wrong-secret")
	s.Require().NoError(err)
	resp.AssertStatus(http.StatusServiceUnavailable)
}

func (s *MaintenanceTestSuite) TestUpCommand_AfterDown() {
	err := facades.Artisan().Call("down")
	s.Require().NoError(err)

	output, err := s.CaptureArtisanOutput("up")
	s.Require().NoError(err)
	s.Contains(output, "The application is up and live now")

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *MaintenanceTestSuite) TestUpCommand_NotInMaintenance() {
	output, err := s.CaptureArtisanOutput("up")
	s.Require().NoError(err)
	s.Contains(output, "The application is not in maintenance mode")
}

func (s *MaintenanceTestSuite) TestDownCommand_WithSecret_Generated() {
	output, err := s.CaptureArtisanOutput("down --with-secret")
	s.Require().NoError(err)
	s.Contains(output, "The application is in maintenance mode now")
	s.Contains(output, "Using secret:")

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertStatus(http.StatusServiceUnavailable)

	match := regexp.MustCompile(`Using secret:\s*(\S+)`).FindStringSubmatch(output)
	s.Require().Len(match, 2)

	secret := ansiEscape.ReplaceAllString(match[1], "")

	resp, err = s.Http(s.T()).Get("/?secret=" + secret)
	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *MaintenanceTestSuite) TestFileDriver_Persistence() {
	storagePath := facades.Storage().Path("framework/maintenance.json")

	s.False(file.Exists(storagePath))

	err := facades.Artisan().Call("down")
	s.Require().NoError(err)
	s.True(file.Exists(storagePath))

	err = facades.Artisan().Call("up")
	s.Require().NoError(err)

	s.False(file.Exists(storagePath))
}

func (s *MaintenanceTestSuite) TestCacheDriver_DownAndUp() {
	facades.Config().Add("app.maintenance", map[string]any{
		"driver": "cache",
		"store":  "",
	})
	s.Require().NoError(facades.App().Restart())

	err := facades.Artisan().Call("down --reason=CacheDriverTest")
	s.Require().NoError(err)

	s.True(facades.Cache().Has("framework:maintenance"))
	s.Contains(facades.Cache().GetString("framework:maintenance"), "CacheDriverTest")

	resp, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertStatus(http.StatusServiceUnavailable)

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal("CacheDriverTest", content)

	err = facades.Artisan().Call("up")
	s.Require().NoError(err)

	s.False(facades.Cache().Has("framework:maintenance"))

	resp, err = s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	facades.Config().Add("app.maintenance", map[string]any{
		"driver": "file",
		"store":  "",
	})
	s.Require().NoError(facades.App().Restart())
}
