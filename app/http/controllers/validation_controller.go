package controllers

import (
	"context"

	"github.com/goravel/framework/contracts/http"
	contractsvalidation "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/validation"
	"github.com/spf13/cast"

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
	Context string                    `json:"context" form:"context"`
	Name    string                    `json:"name" form:"name"`
	Date    *carbon.DateTime          `json:"date" form:"date"`
	Age     int                       `json:"age" form:"age"`
	Items   []requests.ValidationItem `json:"items" form:"items"`
	Meta    map[string]any            `json:"meta" form:"meta"`
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
	validator, err := ctx.Request().Validate(map[string]any{
		"context":      "required",
		"name":         "required",
		"date":         "required|date",
		"items.*.name": "sometimes|required|string",
		"meta":         "sometimes|map",
		"meta.name":    "sometimes|required|string",
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

	response := http.Json{
		"context": user.Context,
		"name":    user.Name,
		"date":    user.Date.ToDateTimeString(),
		"age":     user.Age,
		"items":   user.Items,
		"meta":    user.Meta,
	}

	return ctx.Response().Success().Json(response)
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

	response := http.Json{
		"context": validationCreate.Context,
		"name":    validationCreate.Name,
		"tags":    validationCreate.Tags,
		"scores":  validationCreate.Scores,
		"items":   validationCreate.Items,
		"meta":    validationCreate.Meta,
		"date":    validationCreate.Date.ToDateTimeString(),
		"code":    validationCreate.Code,
		"age":     validationCreate.Age,
	}

	return ctx.Response().Success().Json(response)
}

func (r *ValidationController) Form(ctx http.Context) http.Response {
	ctx.WithValue("ctx", "context")
	validator, err := facades.Validation().Make(ctx, map[string]any{
		"context": ctx.Request().Input("context"),
		"name":    ctx.Request().Input("name"),
		"age":     ctx.Request().Input("age"),
	}, map[string]any{
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
		"age":     user.Age,
	})
}

func (r *ValidationController) Upload(ctx http.Context) http.Response {
	rule := cast.ToString(ctx.Request().Input("rule"))
	message := cast.ToString(ctx.Request().Input("message"))
	if rule == "" {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": "rule is required",
		})
	}

	options := make([]contractsvalidation.Option, 0)
	if message != "" {
		options = append(options, validation.Messages(map[string]string{
			"f." + rule: message,
		}))
	}

	validator, err := ctx.Request().Validate(map[string]any{
		"f": rule,
	}, options...)
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

	return ctx.Response().Success().Json(http.Json{
		"ok": true,
	})
}
