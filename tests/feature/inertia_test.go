package feature

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/testing/http"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

type InertiaTestSuite struct {
	suite.Suite
	tests.TestCase
	originalWD string
}

func TestInertiaTestSuite(t *testing.T) {
	suite.Run(t, &InertiaTestSuite{})
}

func (s *InertiaTestSuite) SetupSuite() {}

// SetupTest switches to the app root before each test. The Inertia root
// template (resources/inertia/app.gohtml) is configured as a CWD-relative path
// and parsed by petaki/inertia-go at render time, while `go test` runs the
// binary from the package dir (tests/feature).
func (s *InertiaTestSuite) SetupTest() {
	wd, err := os.Getwd()
	s.Require().NoError(err)
	s.originalWD = wd
	s.Require().NoError(os.Chdir(path.Base()))
}

func (s *InertiaTestSuite) TearDownTest() {
	s.Require().NoError(os.Chdir(s.originalWD))
}

// inertiaVersion matches the pinned version in config/inertia.go. GET visits
// must echo it back or the Inertia middleware answers 409 + X-Inertia-Location.
const inertiaVersion = "test-version"

func (s *InertiaTestSuite) inertiaRequest() contractshttp.Request {
	return s.Http(s.T()).
		WithHeader("X-Inertia", "true").
		WithHeader("X-Inertia-Version", inertiaVersion)
}

// pageJSON extracts the JSON embedded in the full-page HTML render and
// unmarshals it into a map.
func (s *InertiaTestSuite) pageJSON(resp contractshttp.Response) map[string]any {
	content, err := resp.Content()
	s.Require().NoError(err)
	s.Contains(content, `<script data-page="app" type="application/json">`)
	s.Contains(content, `<div id="app"></div>`)

	start := strings.Index(content, `type="application/json">`) + len(`type="application/json">`)
	end := strings.Index(content[start:], "</script>")
	s.Require().Greater(end, 0, "expected a closing </script> after the page JSON")

	var page map[string]any
	s.Require().NoError(json.Unmarshal([]byte(content[start:start+end]), &page))

	return page
}

// props returns the typed props map from a parsed Inertia page.
func (s *InertiaTestSuite) props(page map[string]any) map[string]any {
	props, ok := page["props"].(map[string]any)
	s.Require().True(ok, "page.props should be an object")

	return props
}

func (s *InertiaTestSuite) TestRenderFullPage() {
	resp, err := s.Http(s.T()).Get("/inertia/home")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	page := s.pageJSON(resp)
	s.Equal("Welcome", page["component"])

	props := s.props(page)
	s.Equal("Goravel + Inertia", props["message"])
	s.Equal(facades.Config().GetString("app.name"), props["appName"])

	auth, ok := props["auth"].(map[string]any)
	s.Require().True(ok, "props.auth should be an object")
	s.Nil(auth["user"])

	s.NotEmpty(props["timestamp"])

	deferred, ok := page["deferredProps"].(map[string]any)
	s.Require().True(ok, "page.deferredProps should be an object")
	defaultGroup, ok := deferred["default"].([]any)
	s.Require().True(ok, "deferredProps.default should be an array")
	s.Contains(defaultGroup, "stats")
}

func (s *InertiaTestSuite) TestRenderInertiaJSON() {
	resp, err := s.inertiaRequest().Get("/inertia/home")
	s.Require().NoError(err)
	resp.AssertSuccessful()
	resp.AssertHeader("X-Inertia", "true")

	page, err := resp.Json()
	s.Require().NoError(err)
	s.Equal("Welcome", page["component"])
}

func (s *InertiaTestSuite) TestVersionMismatchConflict() {
	resp, err := s.Http(s.T()).
		WithHeader("X-Inertia", "true").
		WithHeader("X-Inertia-Version", "stale").
		Get("/inertia/home")
	s.Require().NoError(err)
	resp.AssertConflict()
	s.Contains(resp.Headers().Get("X-Inertia-Location"), "/inertia/home")
}

func (s *InertiaTestSuite) TestFeedMergeProp() {
	resp, err := s.inertiaRequest().Get("/inertia/feed")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	page, err := resp.Json()
	s.Require().NoError(err)
	mergeProps, ok := page["mergeProps"].([]any)
	s.Require().True(ok, "page.mergeProps should be an array")
	s.Contains(mergeProps, "items")
}

func (s *InertiaTestSuite) TestRedirect() {
	resp, err := s.Http(s.T()).Get("/inertia/redirect")
	s.Require().NoError(err)
	resp.AssertFound()

	resp, err = s.Http(s.T()).Delete("/inertia/redirect", nil)
	s.Require().NoError(err)
	resp.AssertStatus(303)
}

func (s *InertiaTestSuite) TestLocation() {
	resp, err := s.Http(s.T()).Get("/inertia/location")
	s.Require().NoError(err)
	resp.AssertFound()

	resp, err = s.inertiaRequest().Get("/inertia/location")
	s.Require().NoError(err)
	resp.AssertConflict()
	s.Contains(resp.Headers().Get("X-Inertia-Location"), "example.com")
}

func (s *InertiaTestSuite) TestShareSessionFlashAndErrors() {
	resp, err := s.inertiaRequest().
		WithSession(map[string]any{
			"success": "Saved!",
			"errors":  map[string]any{"email": "is required"},
		}).
		Get("/inertia/contact")
	s.Require().NoError(err)
	resp.AssertSuccessful()

	page, err := resp.Json()
	s.Require().NoError(err)
	props := s.props(page)

	flash, ok := props["flash"].(map[string]any)
	s.Require().True(ok, "props.flash should be an object")
	s.Equal("Saved!", flash["success"])

	errors, ok := props["errors"].(map[string]any)
	s.Require().True(ok, "props.errors should be an object")
	s.Equal("is required", errors["email"])
}

func (s *InertiaTestSuite) TestStoreContactValidationFails() {
	resp, err := s.Http(s.T()).
		Post("/inertia/contact", strings.NewReader(`{"email":"bad"}`))
	s.Require().NoError(err)
	resp.AssertFound()
}
