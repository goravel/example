package routes

import (
	"strings"

	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
	"goravel/modules/admin/http/controllers"
)

// Web serves the admin page at modules.admin.path and every URL beneath it: the
// admin's own router decides which page is shown, so a refresh on
// /admin/blog/posts works.
func Web(router route.Router) {
	base := "/" + strings.Trim(facades.Config().GetString("modules.admin.path", "admin"), "/")
	controller := controllers.NewDashboardController()

	router.Get(base, controller.Show)
	router.Get(base+"/*page", controller.Show)
}
