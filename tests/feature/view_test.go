package feature

import (
	"testing"

	frameworkerrors "github.com/goravel/framework/errors"
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

// TestMake_RendersAppView verifies a view renders to a string outside a request.
func (s *ViewTestSuite) TestMake_RendersAppView() {
	html, err := facades.View().Make("shared.tmpl", map[string]any{"name": "Goravel"}).Render()
	s.Require().NoError(err)
	s.Equal("<p>app-shared: Goravel</p>", html)
}

// TestMake_RendersPackageView verifies package views registered with
// LoadViewsFrom are reachable through Make.
func (s *ViewTestSuite) TestMake_RendersPackageView() {
	html, err := facades.View().Make("package_only.tmpl", map[string]any{"name": "Goravel"}).Render()
	s.Require().NoError(err)
	s.Equal("<p>package-only: Goravel</p>", html)
}

// TestMake_RendersNestedView verifies partials/layouts resolve outside a request.
// csrf_token is request-bound and absent here; it renders as an empty string.
func (s *ViewTestSuite) TestMake_RendersNestedView() {
	html, err := facades.View().Make("admin/index.tmpl", map[string]any{"name": "Goravel"}).Render()
	s.Require().NoError(err)
	s.Equal("\n  \n<html>\n  <body>\n    <p>I'm the header</p>\n\n  <p>Hello, Goravel</p>\n  <p> CSRF Token:  </p>\n  \n  <p>I'm the footer</p>\n  </body>\n</html>\n\n", html)
}

// TestMake_MatchesHttpResponse verifies Render produces exactly what the driver
// sends for the same view.
func (s *ViewTestSuite) TestMake_MatchesHttpResponse() {
	resp, err := s.Http(s.T()).Get("/shared")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	html, err := facades.View().Make("shared.tmpl", map[string]any{"name": "Goravel"}).Render()
	s.Require().NoError(err)
	s.Equal(content, html)
}

// TestMake_StructData verifies struct and pointer-to-struct data render.
func (s *ViewTestSuite) TestMake_StructData() {
	type greeting struct {
		Name string
	}

	html, err := facades.View().Make("greeting.tmpl", greeting{Name: "Goravel"}).Render()
	s.Require().NoError(err)
	s.Equal("<p>Hello, Goravel</p>", html)

	html, err = facades.View().Make("greeting.tmpl", &greeting{Name: "Goravel"}).Render()
	s.Require().NoError(err)
	s.Equal("<p>Hello, Goravel</p>", html)
}

// TestMake_WithOverridesDataAndShared verifies With overrides passed data and
// that shared data is merged at render time.
func (s *ViewTestSuite) TestMake_WithOverridesDataAndShared() {
	facades.View().Share("Name", "Shared")
	s.T().Cleanup(func() {
		facades.View().Share("Name", nil)
	})

	template := facades.View().Make("greeting.tmpl", map[string]any{"Name": "data"}).
		With("Name", "With")

	s.Equal("greeting.tmpl", template.Name())
	s.Equal(map[string]any{"Name": "With"}, template.Data())

	html, err := template.Render()
	s.Require().NoError(err)
	s.Equal("<p>Hello, With</p>", html)

	// Shared data is used when the template supplies none of its own.
	html, err = facades.View().Make("greeting.tmpl").Render()
	s.Require().NoError(err)
	s.Equal("<p>Hello, Shared</p>", html)
}

// TestMake_MissingView verifies an unknown view fails at Render.
func (s *ViewTestSuite) TestMake_MissingView() {
	_, err := facades.View().Make("missing.tmpl").Render()
	s.ErrorIs(err, frameworkerrors.ViewTemplateNotExist)
}

// TestMake_InvalidData verifies unsupported data is reported at Render.
func (s *ViewTestSuite) TestMake_InvalidData() {
	_, err := facades.View().Make("shared.tmpl", 1).Render()
	s.ErrorIs(err, frameworkerrors.ViewInvalidData)

	_, err = facades.View().Make("shared.tmpl", map[int]string{1: "a"}).Render()
	s.ErrorIs(err, frameworkerrors.ViewInvalidData)
}

// TestFirst verifies First picks the first renderable candidate.
func (s *ViewTestSuite) TestFirst() {
	data := map[string]any{"name": "Goravel"}

	template := facades.View().First([]string{"missing.tmpl", "package_only.tmpl", "shared.tmpl"}, data)
	s.Equal("package_only.tmpl", template.Name())

	html, err := template.Render()
	s.Require().NoError(err)
	s.Equal("<p>package-only: Goravel</p>", html)

	// define_only.tmpl exists but holds only a define block, so it is not
	// renderable under its own name and must be skipped for a later candidate.
	s.True(facades.View().Exists("define_only.tmpl"))
	template = facades.View().First([]string{"define_only.tmpl", "package_only.tmpl"}, data)
	s.Equal("package_only.tmpl", template.Name())

	html, err = template.Render()
	s.Require().NoError(err)
	s.Equal("<p>package-only: Goravel</p>", html)

	// When the first candidate is already renderable it must win.
	template = facades.View().First([]string{"package_only.tmpl", "shared.tmpl"}, data)
	s.Equal("package_only.tmpl", template.Name())

	html, err = template.Render()
	s.Require().NoError(err)
	s.Equal("<p>package-only: Goravel</p>", html)
}

// TestFirst_NoneExist verifies First reports when no candidate exists.
func (s *ViewTestSuite) TestFirst_NoneExist() {
	_, err := facades.View().First([]string{"missing.tmpl", "other.tmpl"}).Render()
	s.ErrorIs(err, frameworkerrors.ViewNoneExist)
}
