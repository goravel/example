package routes

import (
	"strings"

	"goravel/app/facades"
	"goravel/modules/admin/http/controllers"
)

// Web serves the admin page at modules.admin.path and every URL beneath it: the
// admin's own router decides which page is shown, so a refresh on
// /admin/blog/posts works.
func Web() {
	base := "/" + strings.Trim(facades.Config().GetString("modules.admin.path", "admin"), "/")
	controller := controllers.NewDashboardController()

	facades.Route().Get(base, controller.Show)
	facades.Route().Get(base+"/*page", controller.Show)
}
