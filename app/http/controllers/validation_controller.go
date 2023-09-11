package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/requests"
	"goravel/app/models"
)

/*********************************
1. Add route to `/route/web.go`

2. Create UserCreate request and fill it
go run . artisan make:request UserCreate

3. Run Server
air

3. Visit 127.0.0.1:3000/validation/*
 ********************************/

type ValidationController struct {
	//Dependent services
}

func NewValidationController() *ValidationController {
	return &ValidationController{
		//Inject services
	}
}

func (r *ValidationController) Json(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"name": "required",
	})
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}
	if validator.Fails() {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": validator.Errors().All(),
		})
	}

	var user models.User
	if err := validator.Bind(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"name": user.Name,
	})
}

func (r *ValidationController) Request(ctx http.Context) http.Response {
	var userCreate requests.UserCreate
	errors, err := ctx.Request().ValidateRequest(&userCreate)
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}
	if errors != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": errors.All(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"name": userCreate.Name,
	})
}

func (r *ValidationController) Form(ctx http.Context) http.Response {
	validator, err := facades.Validation().Make(map[string]any{
		"name": ctx.Request().Input("name", ""),
	}, map[string]string{
		"name": "required",
	})
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}
	if validator.Fails() {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": validator.Errors().All(),
		})
	}

	var user models.User
	if err := validator.Bind(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"name": user.Name,
	})
}
