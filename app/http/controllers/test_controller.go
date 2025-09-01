package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models/system"
)

type TestController struct {
	// Dependent services
}

func NewTestController() *TestController {
	return &TestController{
		// Inject services
	}
}

func (r *TestController) Test(ctx http.Context) http.Response {
	var user system.User
	if err := facades.Orm().Query().Where("id", 66).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}
