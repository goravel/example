package routes

import (
	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Web() {

	userController := controllers.NewUserController()
	facades.Route().Prefix("/user").Group(func(router route.Router) {

		// auth
		router.Get("/signIn", userController.SignInView).Name("user.signIn")
		router.Post("/signIn", userController.SignInPost).Name("user.signIn.post")
		router.Get("/signUp", userController.SignUpView).Name("user.signup")
		router.Post("/signUp", userController.SignUpPost).Name("user.signUp.post")

		router.Middleware(middleware.AuthUser()).Group(func(router route.Router) {
			router.Get("/", userController.Index).Name("user.index")

		})

	})

	productController := controllers.NewProductController()
	facades.Route().Prefix("/product").Group(func(router route.Router) {
		router.Middleware(middleware.AuthUser()).Get("/", productController.Index).Name("product.index")
	})

	adminController := controllers.NewAdminController()
	facades.Route().Prefix("/admin").Group(func(router route.Router) {

		router.Get("/signIn", adminController.SignInView).Name("admin.signIn")
		router.Post("/signIn", adminController.SignInPost).Name("admin.signIn.post")
		router.Get("/signUp", adminController.SignUpView).Name("admin.signUp")
		router.Post("/signUp", adminController.SignUpPost).Name("admin.signUp.post")

		router.Middleware(middleware.AuthAdmin()).Group(func(router route.Router) {
			router.Get("/", adminController.Index).Name("admin.index")
		})

	})
}
