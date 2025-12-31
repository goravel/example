package controllers

import (
	"context"

	"github.com/goravel/framework/contracts/http"
	contractsvalidation "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/validation"

	"goravel/app/facades"
	"goravel/app/http/requests"
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
	Context string           `json:"context" form:"context"`
	Name    string           `json:"name" form:"name"`
	Date    *carbon.DateTime `json:"date" form:"date"`
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
	ctx.WithValue("ctx", "context")
	validator, err := ctx.Request().Validate(map[string]string{
		"context": "required",
		"name":    "required",
		"date":    "required|date",
	}, validation.PrepareForValidation(func(ctx context.Context, data contractsvalidation.Data) error {
		if c, exist := data.Get("context"); exist {
			// Test getting value from context: ValidationController.Request
			if err := data.Set("context", c.(string)+"_"+ctx.Value("ctx").(string)); err != nil {
				return err
			}
		}

		return nil
	}))
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
		"context": user.Context,
		"name":    user.Name,
		"date":    user.Date.ToDateTimeString(),
	})
}

func (r *ValidationController) Request(ctx http.Context) http.Response {
	// Set context value for testing PrepareForValidation
	ctx.WithValue("ctx", "context")

	var validationCreate requests.ValidationCreate
	errors, err := ctx.Request().ValidateRequest(&validationCreate)
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
		"context": validationCreate.Context,
		"name":    validationCreate.Name,
		"tags":    validationCreate.Tags,
		"scores":  validationCreate.Scores,
		"date":    validationCreate.Date.ToDateTimeString(),
		"code":    validationCreate.Code,
	})
}

func (r *ValidationController) Form(ctx http.Context) http.Response {
	ctx.WithValue("ctx", "context")
	validator, err := facades.Validation().Make(ctx, map[string]any{
		"context": ctx.Request().Input("context"),
		"name":    ctx.Request().Input("name"),
	}, map[string]string{
		"context": "required",
		"name":    "required",
	}, validation.PrepareForValidation(func(ctx context.Context, data contractsvalidation.Data) error {
		if c, exist := data.Get("context"); exist {
			// Test getting value from context: ValidationController.Request
			if err := data.Set("context", c.(string)+"_"+ctx.Value("ctx").(string)); err != nil {
				return err
			}
		}

		return nil
	}))
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
		"context": user.Context,
		"name":    user.Name,
	})
}
