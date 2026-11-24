package blog

import (
	"embed"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/module"

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

func (m *Module) Routes() {
	routes.Web()
	routes.Api()
}

// Config is available as "modules.blog", or through Config().Module("blog").
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
func (m *Module) Views() module.Files {
	return module.FS(views)
}

// Translations are grouped as "blog/<file>", e.g. "blog/messages.title".
func (m *Module) Translations() module.Files {
	return module.FS(lang)
}

// Frontend builds the dashboard script; DashboardScripts has the admin load it.
func (m *Module) Frontend() module.Frontend {
	return module.Frontend{
		Entries: []string{dashboardScript},
	}
}

func (m *Module) DashboardScripts() []string {
	return []string{dashboardScript}
}
