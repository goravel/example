package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

/*********************************
1. Configure DB in .env file
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=goravel
DB_USERNAME=root
DB_PASSWORD=

2. Generate DB
go run . artisan make:migration User
go run . artisan migrate

3. Add route to `/route/web.go`

4. Run Server
air

5. Visit 127.0.0.1:3000/db
 ********************************/

type DBController struct {
	// Dependent services
}

func NewDBController() *DBController {
	return &DBController{
		// Inject services
	}
}

func (r *DBController) Index(ctx http.Context) http.Response {
	// Create user
	if err := facades.Orm().Query().Create(&models.User{
		Name:   "Goravel",
		Avatar: "logo.png",
	}); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	// Fetch all user
	var users []models.User
	if err := facades.Orm().Query().Find(&users); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"length": len(users),
	})
}
