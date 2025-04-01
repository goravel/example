package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support"
	"github.com/spf13/cast"

	"goravel/app/http/controllers"
	"goravel/app/jobs"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		err := facades.Queue().Job(&jobs.Test{}, []any{
			"test", 1,
		}).Dispatch()
		if err != nil {
			facades.Log().Error("Queue job error: %v", err)
		}

		return ctx.Response().View().Make("welcome.tmpl", map[string]any{
			"version": support.Version,
		})
	})

	// Swagger
	swaggerController := controllers.NewSwaggerController()
	facades.Route().Get("/swagger/*any", swaggerController.Index)

	// Single Page Application
	// 1. Add your single page application to `resources/views/*`
	// 2. Add route to `/route/web.go`, needs to contain your home page and static routes
	// 3. Configure nginx based on the /nginx.conf file
	facades.Route().StaticFile("index.html", "./resources/views/index.html")
	facades.Route().Static("css", "./resources/views/css")

	// View Nesting
	// Check the views in `resources/views/admin/*`
	facades.Route().Get("view", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("admin/index.tmpl", map[string]any{
			"name": "Goravel",
		})
	})

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

	facades.Route().Fallback(func(ctx http.Context) http.Response {
		return ctx.Response().String(http.StatusNotFound, "fallback")
	})
}
