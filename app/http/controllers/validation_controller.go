package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"

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

type User struct {
	Name string          `json:"name" form:"name"`
	Date carbon.DateTime `json:"date" form:"date"`
}

type ValidationController struct {
	// Dependent services
}

func NewValidationController() *ValidationController {
	return &ValidationController{
		// Inject services
	}
}

func (r *ValidationController) Json(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"name": "required",
		"date": "required|date",
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

	var user User
	if err := validator.Bind(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"name": user.Name,
		"date": user.Date.ToDateTimeString(),
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
		"name":   userCreate.Name,
		"tags":   userCreate.Tags,
		"scores": userCreate.Scores,
		"date":   userCreate.Date.ToDateTimeString(),
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
