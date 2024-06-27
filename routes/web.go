package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().Json(200, http.Json{
			"Hello": "Goravel",
		})
	})

	// Single Page Application
	// 1. Add your single page application to `resources/views/*`
	// 2. Add route to `/route/web.go`, needs to contain your home page and static routes
	// 3. Configure nginx based on the /nginx.conf file
	facades.Route().Get("web", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("index.html")
	})
	facades.Route().Static("css", "./resources/views/css")

	// View Nesting
	// Check the views in `resources/views/admin/*`
	facades.Route().Get("view", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("admin/index.tmpl", map[string]any{
			"name": "Goravel",
		})
	})

	// Localization
	langController := controllers.NewLangController()
	facades.Route().Middleware(middleware.Lang()).Get("lang", langController.Index)

	// Session
	facades.Route().Prefix("session").Group(func(router route.Router) {
		router.Get("put", func(ctx http.Context) http.Response {
			ctx.Request().Session().Put("name", "Goravel")

			return ctx.Response().Success().Json(http.Json{
				"name": cast.ToString(ctx.Request().Session().Get("name")),
			})
		})
		router.Get("get", func(ctx http.Context) http.Response {
			return ctx.Response().Success().Json(http.Json{
				"name": ctx.Request().Session().Get("name"),
			})
		})
	})

	// Cookie
	facades.Route().Prefix("cookie").Group(func(router route.Router) {
		router.Get("put", func(ctx http.Context) http.Response {
			ctx.Response().Cookie(http.Cookie{
				Name:  "name",
				Value: "Goravel",
			})

			return ctx.Response().Success().String("Set cookie: name=Goravel")
		})
		router.Get("get", func(ctx http.Context) http.Response {
			return ctx.Response().Success().Json(http.Json{
				"name": ctx.Request().Cookie("name"),
			})
		})
	})

	facades.Route().Resource("users", controllers.NewUserController())
}
