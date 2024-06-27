package routes

import (
	"github.com/goravel/framework/facades"
	frameworkmiddleware "github.com/goravel/framework/http/middleware"

	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Api() {
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
	facades.Route().Middleware(frameworkmiddleware.Throttle("login")).Get("/jwt/login", jwtController.Login)
	facades.Route().Middleware(middleware.Jwt()).Get("/jwt", jwtController.Index)

	// Swagger
	swaggerController := controllers.NewSwaggerController()
	facades.Route().Get("/swagger/*any", swaggerController.Index)
}
