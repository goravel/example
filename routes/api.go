package routes

import (
	"github.com/goravel/framework/facades"

	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Api() {
	authController := controllers.NewAuthController()
	facades.Route().Post("auth/login", authController.Login)
	facades.Route().Middleware(middleware.Auth()).Get("auth/info", authController.Info)
}
