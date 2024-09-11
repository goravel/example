package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type AuthController struct {
	// Dependent services
}

func NewAuthController() *AuthController {
	return &AuthController{
		// Inject services
	}
}

func (r *AuthController) Login(ctx http.Context) http.Response {
	// Create a user
	var user models.User
	if err := facades.Orm().Query().FirstOrCreate(&user, models.User{
		Name: ctx.Request().Input("name", "Goravel"),
	}); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	var (
		token string
		err   error
	)

	// Use different guards to login
	if guard := ctx.Request().Header("Guard"); guard == "" {
		token, err = facades.Auth(ctx).LoginUsingID(user.ID)
		if err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
	} else {
		token, err = facades.Auth(ctx).Guard(guard).Login(user)
		if err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
	}

	return ctx.Response().Header("Authorization", "Bearer "+token).Success().Json(http.Json{
		"user": user,
	})
}

func (r *AuthController) Info(ctx http.Context) http.Response {
	var user models.User

	if guard := ctx.Request().Header("Guard"); guard == "" {
		if err := facades.Auth(ctx).User(&user); err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
	} else {
		if err := facades.Auth(ctx).Guard(guard).User(&user); err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}
