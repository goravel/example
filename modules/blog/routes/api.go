package routes

import (
	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
	"goravel/modules/blog/http/controllers"
)

// Api serves /api/blog/posts; the blog's dashboard widget reads the total.
func Api() {
	posts := controllers.NewPostController()

	facades.Route().Prefix("api").Group(func(router route.Router) {
		router.Get("/blog/posts", posts.Index)
		router.Post("/blog/posts", posts.Store)
	})
}
