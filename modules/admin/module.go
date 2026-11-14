package admin

import (
	"embed"
	"io/fs"

	"github.com/goravel/framework/contracts/route"

	"goravel/modules/admin/dashboard"
	"goravel/modules/admin/routes"
)

//go:embed resources/views
var views embed.FS

//go:embed lang
var lang embed.FS

// Module is a Vue admin that other modules fill: they register pages and
// dashboard widgets from a dashboard script (contracts.ProvidesDashboardScripts).
type Module struct{}

func (m *Module) Name() string {
	return dashboard.Name
}

// Order -1 makes the admin the last module, so its routes come after the others'.
func (m *Module) Order() int {
	return -1
}

// Config is available as "modules.admin": path is where the admin answers.
func (m *Module) Config() map[string]any {
	return map[string]any{
		"path": "admin",
	}
}

func (m *Module) Routes(router route.Router) {
	routes.Web(router)
}

// Views are named "admin/<file>". The application already has views under
// resources/views/admin, so this module's page is "admin/dashboard.tmpl".
func (m *Module) Views() fs.FS {
	sub, _ := fs.Sub(views, "resources/views")

	return sub
}

func (m *Module) Translations() fs.FS {
	sub, _ := fs.Sub(lang, "lang")

	return sub
}

func (m *Module) ViteEntries() []string {
	return []string{dashboard.Main}
}
