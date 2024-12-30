package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support"

	"goravel/app/http/controllers"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
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
}
