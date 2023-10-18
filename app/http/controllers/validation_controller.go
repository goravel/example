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

4. Visit 127.0.0.1:3000/validation/*
4.1 curl --location --request POST 'http://127.0.0.1:3000/validation/json'
4.2 curl --location --request POST 'http://127.0.0.1:3000/validation/json' --header 'Content-Type: application/json' --data-raw '{"name": ""}'
4.3 curl --location --request POST 'http://127.0.0.1:3000/validation/json' --header 'Content-Type: application/json' --data-raw '{"name": "goravel"}'
4.4 curl --location --request POST 'http://127.0.0.1:3000/validation/request'
4.5 curl --location --request POST 'http://127.0.0.1:3000/validation/request' --header 'Content-Type: application/json' --data-raw '{"name": ""}'
4.6 curl --location --request POST 'http://127.0.0.1:3000/validation/request' --header 'Content-Type: application/json' --data-raw '{"name": "goravel"}'
4.7 curl --location --request POST 'http://127.0.0.1:3000/validation/form'
4.8 curl --location --request POST 'http://127.0.0.1:3000/validation/form' --header 'Content-Type: multipart/form-data' --form 'name=""'
4.9 curl --location --request POST 'http://127.0.0.1:3000/validation/form' --header 'Content-Type: multipart/form-data' --form 'name="goravel"'
 ********************************/

type ValidationController struct {
	// Dependent services
}

func NewValidationController() *ValidationController {
	return &ValidationController{
		// Inject services
	}
}

func (r *ValidationController) Json(ctx http.Context) {
	validator, err := ctx.Request().Validate(map[string]string{
		"name": "required",
	})
	if err != nil {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
		return
	}
	if validator.Fails() {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": validator.Errors().All(),
		})
		return
	}

	var user models.User
	if err = validator.Bind(&user); err != nil {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
		return
	}

	ctx.Response().Success().Json(http.Json{
		"name": user.Name,
	})
}

func (r *ValidationController) Request(ctx http.Context) {
	var userCreate requests.UserCreate
	errors, err := ctx.Request().ValidateRequest(&userCreate)
	if err != nil {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
		return
	}
	if errors != nil {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": errors.All(),
		})
		return
	}

	ctx.Response().Success().Json(http.Json{
		"name": userCreate.Name,
	})
}

func (r *ValidationController) Form(ctx http.Context) {
	validator, err := facades.Validation().Make(map[string]any{
		"name": ctx.Request().Form("name", ""),
	}, map[string]string{
		"name": "required",
	})
	if err != nil {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
		return
	}
	if validator.Fails() {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": validator.Errors().All(),
		})
		return
	}

	var user models.User
	if err = validator.Bind(&user); err != nil {
		ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
		return
	}

	ctx.Response().Success().Json(http.Json{
		"name": user.Name,
	})
}
