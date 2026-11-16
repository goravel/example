package routes

import (
	"github.com/goravel/framework/contracts/route"

	"goravel/modules/blog/http/controllers"
)

// Api is what the blog's admin pages call: /api/blog/posts.
func Api(router route.Router) {
	posts := controllers.NewPostController()

	router.Prefix("api").Group(func(router route.Router) {
		router.Get("/blog/posts", posts.Index)
		router.Post("/blog/posts", posts.Store)
	})
}
