package blog

import (
	"embed"
	"io/fs"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/route"

	admincontracts "goravel/modules/admin/contracts"
	"goravel/modules/blog/database/migrations"
	"goravel/modules/blog/routes"
)

//go:embed resources/views
var views embed.FS

//go:embed lang
var lang embed.FS

// dashboardScript registers the blog's admin pages and its stat card.
const dashboardScript = "resources/js/dashboard.ts"

// The admin finds the blog by this interface; the assertion only documents it.
var _ admincontracts.ProvidesDashboardScripts = (*Module)(nil)

type Module struct{}

func (m *Module) Name() string {
	return "blog"
}

func (m *Module) Routes(router route.Router) {
	routes.Web(router)
}

func (m *Module) ApiRoutes(router route.Router) {
	routes.Api(router)
}

// Config is available as "modules.blog".
func (m *Module) Config() map[string]any {
	return map[string]any{
		"per_page": 10,
	}
}

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20261114055445CreateBlogPostsTable{},
	}
}

// Views are named "blog/<file>", e.g. "blog/index.tmpl".
func (m *Module) Views() fs.FS {
	sub, _ := fs.Sub(views, "resources/views")

	return sub
}

// Translations are grouped as "blog/<file>", e.g. "blog/messages.title".
func (m *Module) Translations() fs.FS {
	sub, _ := fs.Sub(lang, "lang")

	return sub
}

// ViteEntries builds the dashboard script; DashboardScripts has the admin load it.
func (m *Module) ViteEntries() []string {
	return []string{dashboardScript}
}

func (m *Module) DashboardScripts() []string {
	return []string{dashboardScript}
}
