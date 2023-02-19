package controllers

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type UserController struct {
	//Dependent services
}

func NewUserController() *UserController {
	return &UserController{
		//Inject services
	}
}

func (r *UserController) Show(ctx contractshttp.Context) {
	// Create user
	if err := facades.Orm.Query().Create(&models.User{
		Name:   "Goravel",
		Avatar: "logo.png",
	}); err != nil {
		ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{
			"error": err.Error(),
		})
		return
	}

	// Fetch all user
	var users []models.User
	if err := facades.Orm.Query().Find(&users); err != nil {
		ctx.Response().Json(http.StatusInternalServerError, contractshttp.Json{
			"error": err.Error(),
		})
		return
	}

	ctx.Response().Success().Json(contractshttp.Json{
		"length": len(users),
	})
}
