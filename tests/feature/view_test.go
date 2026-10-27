package feature

import (
	"fmt"
	"testing"

	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// fixedCSRFToken is written into the session as `_token`; the driver copies
// session.Token() into the shared csrf_token / csrf_field values, so the
// rendered output is deterministic.
const fixedCSRFToken = "csrf-field-test-token"

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

func (s *ViewTestSuite) TestRender_ViewFuncs() {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "should render an application view function",
			path: "/view-funcs/app",
			want: "<p>app-func: app-hello Goravel</p>",
		},
		{
			name: "should render a package view function",
			path: "/view-funcs/package",
			want: "<p>package-func: package-hello Goravel</p>",
		},
		{
			// GreetingFuncName is registered by both the package provider and the
			// application; the application's registration is applied last.
			name: "should prefer the application function on a name collision",
			path: "/view-funcs/override",
			want: "<p>greeting: app-hello Goravel</p>",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp, err := s.Http(s.T()).Get(tt.path)
			s.Require().NoError(err)
			resp.AssertSuccessful()

			content, err := resp.Content()
			s.Require().NoError(err)
			s.Contains(content, tt.want)
		})
	}
}

// TestFuncs_AppAndPackage verifies View.Funcs exposes the merged registry and
// that the application's function won the collision.
func (s *ViewTestSuite) TestFuncs_AppAndPackage() {
	funcs := facades.View().Funcs()
	s.Require().NotNil(funcs)

	// Membership only: a new framework-registered function must not fail this.
	for _, name := range []string{"app_greeting", "greeting", "package_greeting"} {
		s.Contains(funcs, name)
	}

	greeting, ok := funcs["greeting"].(func(string) string)
	s.Require().True(ok)
	s.Equal("app-hello Goravel", greeting("Goravel"))
}

// TestFuncs_ReturnedMapIsACopy verifies View.Funcs hands out a copy, so a caller
// cannot mutate the registry through it.
func (s *ViewTestSuite) TestFuncs_ReturnedMapIsACopy() {
	funcs := facades.View().Funcs()
	s.Require().NotNil(funcs)
	funcs["__probe"] = func(string) string { return "mutated" }

	_, ok := facades.View().Funcs()["__probe"]
	s.False(ok)
}

// TestRender_CsrfField verifies `{{ .csrf_field }}` from gin#258 / fiber#313:
// the value is template.HTML, so html/template writes the <input> tag unescaped,
// while the token itself is HTML-escaped before it is embedded in the attribute.
func (s *ViewTestSuite) TestRender_CsrfField() {
	const (
		hostileToken = `"><script>alert(1)</script>`
		escapedToken = "&#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt;"
		body         = `<p>csrf_token=%s</p><input type="hidden" name="_token" value="%s" />`
	)

	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "should render csrf_field as unescaped html",
			token:    fixedCSRFToken,
			expected: fmt.Sprintf(body, fixedCSRFToken, fixedCSRFToken),
		},
		{
			name:     "should escape the token in csrf_field",
			token:    hostileToken,
			expected: fmt.Sprintf(body, escapedToken, escapedToken),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp, err := s.Http(s.T()).
				WithSession(map[string]any{"_token": tt.token}).
				Get("/view-csrf/field")
			s.Require().NoError(err)
			resp.AssertSuccessful()

			content, err := resp.Content()
			s.Require().NoError(err)
			s.Equal(tt.expected, content)
		})
	}
}

// TestRender_StringMapDataWithCsrfField locks in the fillShared guard from
// gin#258 / fiber#313: with a session present, csrf_field (template.HTML) is not
// assignable to a map[string]string element, so the driver must skip it instead
// of panicking in reflect.Value.SetMapIndex, while csrf_token (string) is still
// merged in. The pre-bump pins had neither csrf_field nor the guard, so this case
// was already green there: it pins the guard going forward rather than
// reproducing a 500 from the old pins.
func (s *ViewTestSuite) TestRender_StringMapDataWithCsrfField() {
	resp, err := s.Http(s.T()).
		WithSession(map[string]any{"_token": fixedCSRFToken}).
		Get("/view-csrf/string-map")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Equal(
		fmt.Sprintf("<p>string-map: Goravel</p><p>csrf_token=%s</p>", fixedCSRFToken),
		content,
	)
}
