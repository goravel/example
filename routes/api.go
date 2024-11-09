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

	testController := controllers.NewTestController()
	facades.Route().Get("test/index", testController.Index)
	facades.Route().Get("test/show/{id}", testController.Show)
	facades.Route().Post("test/create", testController.Create)
	facades.Route().Put("test/update/{id}", testController.Update)
	facades.Route().Delete("test/delete/{id}", testController.Delete)
	facades.Route().Get("test/custom-header", testController.CustomHeader)
	facades.Route().Get("test/empty-response", testController.EmptyResponse)
	facades.Route().Middleware(middleware.Auth()).Post("test/authorization", testController.Index)
}
