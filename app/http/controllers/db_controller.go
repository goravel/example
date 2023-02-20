package controllers

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

/*********************************
1. Generate DB
go run . artisan make:migration User
go run . artisan migrate

2. Generate Controller

3. Add route to `/route/web.go`

4. Run Server
air

5. Visit 127.0.0.1:3000/db
 ********************************/

type DBController struct {
	//Dependent services
}

func NewDBController() *DBController {
	return &DBController{
		//Inject services
	}
}

func (r *DBController) Index(ctx contractshttp.Context) {
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
