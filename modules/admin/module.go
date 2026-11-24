package admin

import (
	"embed"

	"github.com/goravel/framework/contracts/module"

	"goravel/modules/admin/dashboard"
	"goravel/modules/admin/routes"
)

//go:embed resources/views
var views embed.FS

//go:embed lang
var lang embed.FS

// Module is an admin page that other modules fill with widgets from their own
// dashboard scripts (contracts.ProvidesDashboardScripts).
type Module struct{}

func (m *Module) Name() string {
	return dashboard.Name
}

// Config is available as "modules.admin": path is where the admin answers.
func (m *Module) Config() map[string]any {
	return map[string]any{
		"path": "admin",
	}
}

// LateRoutes runs after every module's routes.
func (m *Module) LateRoutes() {
	routes.Web()
}

// Views are named "admin/<file>". The application already has views under
// resources/views/admin, so this module's page is "admin/dashboard.tmpl".
func (m *Module) Views() module.Files {
	return module.Files{FS: views}
}

func (m *Module) Translations() module.Files {
	return module.Files{FS: lang}
}

// Frontend has Vite build the admin's entry, which no view names.
func (m *Module) Frontend() module.Frontend {
	return module.Frontend{
		Entries: []string{dashboard.Main},
	}
}
