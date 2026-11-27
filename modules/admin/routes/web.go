package routes

import (
	"strings"

	"goravel/app/facades"
	"goravel/modules/admin/http/controllers"
)

// Web serves the admin page at modules.admin.path.
func Web() {
	base := "/" + strings.Trim(facades.Config().Module("admin").GetString("path", "admin"), "/")
	facades.Route().Get(base, controllers.NewDashboardController().Show)
}
