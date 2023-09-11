package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpswagger "github.com/swaggo/http-swagger"

	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().Json(200, http.Json{
			"Hello": "Goravel",
		})
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

	// JWT
	jwtController := controllers.NewJwtController()
	facades.Route().Get("/jwt/login", jwtController.Login)
	facades.Route().Middleware(middleware.Jwt()).Get("/jwt", jwtController.Index)

	// Swagger
	swaggerController := controllers.NewSwaggerController()
	facades.Route().Get("/swagger", swaggerController.Index)
	facades.Route().StaticFile("/swagger.json", "./docs/swagger.json")
	facades.Route().Get("/swagger/*any", func(ctx http.Context) http.Response {
		handler := httpswagger.Handler(httpswagger.URL("http://localhost:3000/swagger.json"))
		handler(ctx.Response().Writer(), ctx.Request().Origin())

		return nil
	})
}
