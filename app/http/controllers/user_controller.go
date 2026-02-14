package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"goravel/app/facades"
	"goravel/app/http/requests"
	"goravel/app/models"
)

type UserController struct {
	errorCounter metric.Int64Counter
}

func NewUserController() *UserController {
	meter := facades.Telemetry().MeterProvider().Meter("user_controller")

	// We use an Int64Counter for counting discrete error events
	errCounter, _ := meter.Int64Counter(
		"user_controller_errors_total",
		metric.WithDescription("Total number of errors in user controller"),
	)

	return &UserController{
		errorCounter: errCounter,
	}
}

func (r *UserController) Index(ctx http.Context) http.Response {
	var users []models.User
	if err := facades.Orm().Query().Get(&users); err != nil {
		r.recordError(ctx, "Index", "db_error")
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"users": users,
	})
}

func (r *UserController) Show(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Store(ctx http.Context) http.Response {
	var userCreate requests.UserCreate
	errors, err := ctx.Request().ValidateRequest(&userCreate)
	if err != nil {
		r.recordError(ctx, "Store", "validation_error")
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}
	if errors != nil {
		r.recordError(ctx, "Store", "validation_error")
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": errors.All(),
		})
	}

	user := models.User{
		Name:   userCreate.Name,
		Avatar: userCreate.Avatar,
		Alias:  userCreate.Alias,
		Mail:   userCreate.Mail,
		Tags:   userCreate.Tags,
	}
	if err := facades.Orm().Query().Create(&user); err != nil {
		r.recordError(ctx, "Store", "db_create_error")
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Update(ctx http.Context) http.Response {
	if _, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Update(models.User{
		Name: ctx.Request().Input("name"),
	}); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Destroy(ctx http.Context) http.Response {
	result, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Delete(&models.User{})
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"rows_affected": result.RowsAffected,
	})
}

func (r *UserController) recordError(ctx http.Context, method string, errType string) {
	r.errorCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", method),      // e.g., "Index", "Store"
		attribute.String("error_type", errType), // e.g., "db_error", "validation_error"
	))
}
