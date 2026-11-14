package feature

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/modules/blog/models"
	"goravel/tests"
)

// ModularTestSuite shows how the framework reads the modules listed in
// bootstrap/modules.go: modules/blog (routes, API routes, config, migrations,
// views, translations, a Vite entry) and modules/admin, which the blog plugs
// into through the admin's own capability (see admin_dashboard_test.go).
type ModularTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestModularTestSuite(t *testing.T) {
	suite.Run(t, &ModularTestSuite{})
}

// The blog's table comes from its Migrations(), run by TestMain's migrate.
func (s *ModularTestSuite) SetupTest() {
	s.RefreshDatabase()
}

func (s *ModularTestSuite) TestRegistry() {
	registry := facades.Modular()

	// bootstrap/modules.go lists admin first, but its Order() is -1: last.
	var names []string
	for _, module := range registry.All() {
		names = append(names, module.Name())
	}
	s.Equal([]string{"blog", "admin"}, names)

	// The directory is derived from the Go package, goravel/modules/blog.
	dir, ok := registry.Path("blog")
	s.True(ok)
	s.Equal("modules/blog", dir)

	blog, ok := registry.Find("blog")
	s.Require().True(ok)
	s.Equal(map[string]bool{
		"services":     false,
		"config":       true,
		"middleware":   false,
		"routes":       true,
		"api":          true,
		"migrations":   true,
		"seeders":      false,
		"commands":     false,
		"jobs":         false,
		"events":       false,
		"schedule":     false,
		"views":        true,
		"translations": true,
		"vite-entries": true,
		"vite-aliases": false,
		"static":       false,
		"sources":      false,
	}, registry.Capabilities(blog))

	s.False(registry.Has("comments"))
}

func (s *ModularTestSuite) TestConfigAndTranslations() {
	// Config() lands under "modules.<name>".
	s.Equal(10, facades.Config().GetInt("modules.blog.per_page"))
	s.Equal("admin", facades.Config().GetString("modules.admin.path"))

	// lang/en/messages.json of a module is the group "<name>/messages".
	s.Equal("Blog", facades.Lang(context.Background()).Get("blog/messages.title"))
	s.Equal("Admin", facades.Lang(context.Background()).Get("admin/messages.title"))
}

func (s *ModularTestSuite) TestModularListCommand() {
	output, err := s.CaptureArtisanOutput("--no-ansi modular:list")
	s.Require().NoError(err)

	s.Contains(output, "blog modules/blog")
	s.Contains(output, "local, order 0")
	s.Contains(output, "api, config, migrations, routes, translations, views, vite-entries")
	s.Contains(output, "vite entries: resources/js/dashboard.ts")
	s.Contains(output, "admin modules/admin")
	s.Contains(output, "local, order -1")
	s.Less(strings.Index(output, "blog modules/blog"), strings.Index(output, "admin modules/admin"))
}

// TestViteInputs pins what goravel-vite-plugin builds for the modules: each
// module's ViteEntries joined onto its directory, in module order.
func (s *ModularTestSuite) TestViteInputs() {
	output, err := s.CaptureArtisanOutput("vite:inputs")
	s.Require().NoError(err)

	s.Contains(output, `"modules":[`+
		`{"name":"blog","path":"modules/blog","entries":["modules/blog/resources/js/dashboard.ts"],"aliases":{}},`+
		`{"name":"admin","path":"modules/admin","entries":["modules/admin/resources/js/main.ts"],"aliases":{}}]`)
}

func (s *ModularTestSuite) TestBlogApi() {
	s.Run("should refuse a post without title and body", func() {
		resp, err := s.Http(s.T()).Post("/api/blog/posts", strings.NewReader(`{}`))
		s.Require().NoError(err)
		resp.AssertUnprocessableEntity().AssertExactJson(map[string]any{
			"errors": map[string]any{
				"title": map[string]any{"required": "The title field is required."},
				"body":  map[string]any{"required": "The body field is required."},
			},
		})
	})

	s.Run("should create a post", func() {
		resp, err := s.Http(s.T()).
			WithHeader("Content-Type", "application/json").
			Post("/api/blog/posts", strings.NewReader(`{"title":"Hello modules","body":"First post"}`))
		s.Require().NoError(err)
		resp.AssertCreated()

		body, err := resp.Json()
		s.Require().NoError(err)
		post := body["data"].(map[string]any)
		s.Equal("Hello modules", post["title"])
		s.Equal("First post", post["body"])
		s.NotZero(post["id"])

		count, err := facades.Orm().Query().Model(&models.Post{}).Count()
		s.Require().NoError(err)
		s.Equal(int64(1), count)
	})

	s.Run("should list the latest posts with the total", func() {
		s.Require().NoError(facades.Orm().Query().Create(&models.Post{Title: "Second", Body: "Body"}))

		resp, err := s.Http(s.T()).Get("/api/blog/posts")
		s.Require().NoError(err)
		resp.AssertOk()

		body, err := resp.Json()
		s.Require().NoError(err)
		s.Equal(float64(2), body["total"])

		var titles []string
		for _, post := range body["data"].([]any) {
			titles = append(titles, post.(map[string]any)["title"].(string))
		}
		s.Equal([]string{"Second", "Hello modules"}, titles)
	})
}

// TestBlogPage renders a module view ("blog/index.tmpl") with a module translation.
func (s *ModularTestSuite) TestBlogPage() {
	s.Require().NoError(facades.Orm().Query().Create(&models.Post{Title: "Hello modules", Body: "First post"}))

	resp, err := s.Http(s.T()).Get("/blog")
	s.Require().NoError(err)
	resp.AssertOk().AssertSeeInOrder([]string{"<title>Blog</title>", "<h2>Hello modules</h2>", "<p>First post</p>"}, false)
}
