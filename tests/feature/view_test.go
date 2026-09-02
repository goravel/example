package feature

import (
	"testing"

	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

type ViewTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestViewTestSuite(t *testing.T) {
	suite.Run(t, &ViewTestSuite{})
}

func (s *ViewTestSuite) SetupSuite() {}

func (s *ViewTestSuite) SetupTest() {}

func (s *ViewTestSuite) TearDownTest() {}

func (s *ViewTestSuite) TestRegisteredViews() {
	paths := facades.View().RegisteredViews()
	s.Contains(paths, path.Base("packages", "viewtest", "views"))
}

func (s *ViewTestSuite) TestExists_PackageOnly() {
	s.True(facades.View().Exists("package_only.tmpl"))
}

func (s *ViewTestSuite) TestExists_Shared() {
	s.True(facades.View().Exists("shared.tmpl"))
}

func (s *ViewTestSuite) TestRender_PackageOnly() {
	resp, err := s.Http(s.T()).Get("/package-only")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Contains(content, "<p>package-only: Goravel</p>")
}

// TestRender_SharedViewOverride verifies that resources/views/shared.tmpl
// overrides packages/viewtest/views/shared.tmpl. Both shared.tmpl files are
// now named ({{ define "shared.tmpl" }}), so both drivers skip the package
// template when the app already defines the name: the app view wins.
func (s *ViewTestSuite) TestRender_SharedViewOverride() {
	resp, err := s.Http(s.T()).Get("/shared")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Contains(content, "<p>app-shared: Goravel</p>")
}
