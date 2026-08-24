package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/http/middleware"
	"github.com/goravel/framework/support"
	"github.com/spf13/cast"

	"goravel/app/facades"
	"goravel/app/http/controllers"
	appmiddleware "goravel/app/http/middleware"
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
	facades.Route().StaticFile("index.html", "./resources/views/index.html")
	facades.Route().Static("css", "./resources/views/css")

	// Inertia static assets: the committed root template links /public/favicon.png
	// and the vite helper emits /build/... URLs from the production manifest.
	facades.Route().Static("public", "./public")
	facades.Route().Static("build", "./public/build")

	// View Nesting
	// Check the views in `resources/views/admin/*`
	facades.Route().Middleware(middleware.VerifyCsrfToken()).Get("view", func(ctx http.Context) http.Response {
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

	// Package View Test: template only exists in package viewtest/views/
	facades.Route().Get("package-only", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("package_only.tmpl", map[string]any{
			"name": "Goravel",
		})
	})

	// Package View Test: template overridden in resources/views/
	facades.Route().Get("shared", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("shared.tmpl", map[string]any{
			"name": "Goravel",
		})
	})

	facades.Route().Fallback(func(ctx http.Context) http.Response {
		return ctx.Response().String(http.StatusNotFound, "fallback")
	})

	// Test Broadcasting
	broadcastController := controllers.NewBroadcastController()
	facades.Route().Post("broadcasting/dispatch", broadcastController.Dispatch)
	facades.Route().StaticFile("broadcast.html", "./resources/views/broadcast.html")
	facades.Route().StaticFile("echo.html", "./resources/views/echo.html")

	// Inertia (session-based web routes; StartSession is already global)
	facades.Route().Middleware(
		appmiddleware.Inertia(),
	).Prefix("inertia").Group(func(router route.Router) {
		inertiaController := controllers.NewInertiaController()
		router.Get("home", inertiaController.Home)
		router.Get("feed", inertiaController.Feed)
		router.Get("contact", inertiaController.Contact)
		router.Post("contact", inertiaController.StoreContact)
		router.Get("redirect", inertiaController.RedirectTo)
		router.Delete("redirect", inertiaController.RedirectTo)
		router.Get("location", inertiaController.LocationTo)
	})
}
