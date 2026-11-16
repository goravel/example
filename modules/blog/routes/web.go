package routes

import (
	"goravel/app/facades"
	"goravel/modules/blog/http/controllers"
)

func Web() {
	facades.Route().Get("/blog", controllers.NewBlogController().Index)
}
