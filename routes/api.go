package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	httpmiddleware "github.com/goravel/framework/http/middleware"

	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Api() {
	// Auth
	authController := controllers.NewAuthController()
	facades.Route().Prefix("jwt").Group(func(route route.Router) {
		route.Post("login", authController.LoginByJwt)
		route.Middleware(middleware.Jwt()).Get("info", authController.InfoByJwt)
	})

	facades.Route().Prefix("session").Group(func(route route.Router) {
		route.Post("login", authController.LoginBySession)
		route.Middleware(middleware.Session()).Get("info", authController.InfoBySession)
	})

	// DB
	dbController := controllers.NewDBController()
	facades.Route().Get("/db", dbController.Index)

	// Websocket
	websocketController := controllers.NewWebsocketController()
	facades.Route().Get("/ws", websocketController.Server)

	// Validation
	validationController := controllers.NewValidationController()
	facades.Route().Post("/validation/json", validationController.Json)
	facades.Route().Post("/validation/request", validationController.Request)
	facades.Route().Post("/validation/form", validationController.Form)

	// Localization
	langController := controllers.NewLangController()
	facades.Route().Middleware(middleware.Lang()).Get("lang", langController.Index)

	// Test Rate Limiter
	facades.Route().Middleware(httpmiddleware.Throttle("ip")).Get("/throttle", func(ctx http.Context) http.Response {
		return ctx.Response().Success().String("success")
	})
}
