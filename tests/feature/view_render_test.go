package feature

import (
	"strings"
	"testing"

	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// ViewRenderTestSuite covers View.Make and View.First (framework#1594), the
// rendering path that returns a string and needs no http.Context. The HTTP side
// of the same views is covered by ViewTestSuite in view_test.go.
type ViewRenderTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestViewRenderTestSuite(t *testing.T) {
	suite.Run(t, &ViewRenderTestSuite{})
}

// normalizeLineEndings keeps the LF expectations portable: templates are read
// from disk byte-for-byte and the repo has no .gitattributes, so a Windows
// checkout can hand back CRLF.
func normalizeLineEndings(html string) string {
	return strings.ReplaceAll(html, "\r\n", "\n")
}

// TestRender_RendersOutsideARequest pins that a template renders with no request
// in scope: request-bound values such as csrf_token come out empty, while nested
// templates and path-named files (no {{ define }}) still resolve.
func (s *ViewRenderTestSuite) TestRender_RendersOutsideARequest() {
	s.Run("should render a file without define blocks by its path", func() {
		html, err := facades.View().Make("errors/503.tmpl").Render()
		s.Require().NoError(err)
		s.Equal("Service Unavailable - Maintenance Mode\n", normalizeLineEndings(html))
	})

	// Only what the case is about: the nested templates resolved and the empty
	// CSRF slot, not the admin views' own whitespace.
	s.Run("should leave request values empty outside a request", func() {
		html, err := facades.View().Make("admin/index.tmpl", map[string]any{"name": "Goravel"}).Render()
		s.Require().NoError(err)

		s.Contains(html, "<p>I'm the header</p>")
		s.Contains(html, "<p>Hello, Goravel</p>")
		s.Contains(html, "<p> CSRF Token:  </p>")
		s.Contains(html, "<p>I'm the footer</p>")
	})
}

// TestMake_DataNameAndWith also renders a package view: package_only.tmpl is
// registered through LoadViewsFrom and resources/views has no file of that name.
func (s *ViewRenderTestSuite) TestMake_DataNameAndWith() {
	tmpl := facades.View().Make("package_only.tmpl", map[string]any{"name": "data"})
	s.Equal("package_only.tmpl", tmpl.Name())
	s.Equal(map[string]any{"name": "data"}, tmpl.Data())

	html, err := tmpl.With("name", "Goravel").Render()
	s.Require().NoError(err)
	s.Equal("<p>package-only: Goravel</p>", html)
	s.Equal(map[string]any{"name": "Goravel"}, tmpl.Data())
}

// TestMake_AppViewWinsOverPackageView pins the same precedence /shared exercises
// over HTTP: resources/views/shared.tmpl and packages/viewtest/views/shared.tmpl
// both define "shared.tmpl", the app file is compiled first, so the package file
// is skipped whole and Make renders the app one.
func (s *ViewRenderTestSuite) TestMake_AppViewWinsOverPackageView() {
	s.True(facades.View().Exists("shared.tmpl"))

	html, err := facades.View().Make("shared.tmpl", map[string]any{"name": "Goravel"}).Render()
	s.Require().NoError(err)
	s.Equal("<p>app-shared: Goravel</p>", html)
}

// TestFirst_SkipsMissingAndDefineOnlyViews covers why First cannot just ask
// Exists: packages/viewtest/views/partials/nav.tmpl is a real file that only
// defines "nav.tmpl", so it exists without being renderable under its own path
// and First has to carry on to the next candidate.
func (s *ViewRenderTestSuite) TestFirst_SkipsMissingAndDefineOnlyViews() {
	s.True(facades.View().Exists("partials/nav.tmpl"))

	html, err := facades.View().Make("partials/nav.tmpl").Render()
	s.Require().ErrorIs(err, frameworkerrors.ViewTemplateNotExist)
	s.Empty(html)

	html, err = facades.View().Make("nav.tmpl").Render()
	s.Require().NoError(err)
	s.Equal("<nav>navigation</nav>", html)

	tmpl := facades.View().First(
		[]string{"missing.tmpl", "partials/nav.tmpl", "package_only.tmpl"},
		map[string]any{"name": "Goravel"},
	)
	s.Equal("package_only.tmpl", tmpl.Name())

	html, err = tmpl.Render()
	s.Require().NoError(err)
	s.Equal("<p>package-only: Goravel</p>", html)
}

// TestFirst_NoRenderableView pins the empty result shape: no name, no data, and
// Render reporting errors.ViewNoneExist instead of a template error.
func (s *ViewRenderTestSuite) TestFirst_NoRenderableView() {
	tmpl := facades.View().First([]string{"missing.tmpl", "partials/nav.tmpl"})
	s.Empty(tmpl.Name())
	s.Equal(map[string]any{}, tmpl.Data())

	html, err := tmpl.Render()
	s.Equal("", html)
	s.Require().ErrorIs(err, frameworkerrors.ViewNoneExist)
}

// TestFirst_InvalidData pins the other half of the no-match branch: unusable
// data is still reported when no candidate rendered, and the message names the
// candidates it was meant for.
func (s *ViewRenderTestSuite) TestFirst_InvalidData() {
	html, err := facades.View().First(
		[]string{"missing.tmpl", "partials/nav.tmpl"},
		map[int]string{1: "one"},
	).Render()
	s.Empty(html)
	s.Require().ErrorIs(err, frameworkerrors.ViewInvalidData)
	s.Contains(err.Error(), "partials/nav.tmpl")
}

// TestMake_InvalidData pins that unusable data is reported by Render, not by
// Make, and that the message names the view it was meant for.
func (s *ViewRenderTestSuite) TestMake_InvalidData() {
	tests := []struct {
		name string
		data any
	}{
		{name: "should reject a map with non-string keys", data: map[int]string{1: "one"}},
		{name: "should reject data that is not a map or a struct", data: 1},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			html, err := facades.View().Make("shared.tmpl", tt.data).Render()
			s.Equal("", html)
			s.Require().ErrorIs(err, frameworkerrors.ViewInvalidData)
			s.Contains(err.Error(), "shared.tmpl")
		})
	}
}

// TestRender_MergesSharedData covers Share + Make: shared data is visible when the
// template carries none, template data wins over it, and rendering does not write
// back into the shared map. Shared data lasts for the whole process, so the
// previous value goes back afterwards; when there was none, the key is left
// holding nil because View cannot remove shared data.
func (s *ViewRenderTestSuite) TestRender_MergesSharedData() {
	previous := facades.View().Shared("name")
	facades.View().Share("name", "shared-name")
	s.T().Cleanup(func() { facades.View().Share("name", previous) })

	tmpl := facades.View().Make("shared.tmpl")
	s.Equal(map[string]any{}, tmpl.Data())

	html, err := tmpl.Render()
	s.Require().NoError(err)
	s.Equal("<p>app-shared: shared-name</p>", html)

	html, err = facades.View().Make("shared.tmpl", map[string]any{"name": "data-name"}).Render()
	s.Require().NoError(err)
	s.Equal("<p>app-shared: data-name</p>", html)

	s.Equal("shared-name", facades.View().Shared("name"))
}
