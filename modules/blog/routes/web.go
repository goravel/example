package routes

import (
	"github.com/goravel/framework/contracts/route"

	"goravel/modules/blog/http/controllers"
)

func Web(router route.Router) {
	router.Get("/blog", controllers.NewBlogController().Index)
}
