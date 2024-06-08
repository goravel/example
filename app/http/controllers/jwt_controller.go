package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

/*********************************
1. Generate JWT secret
go run . artisan jwt:secret

2. Generate Middleware
go run . artisan make:middleware Jwt

3. Add route to `/route/web.go`

4. Run Server
air

5. Visit 127.0.0.1:3000/jwt/login to get token
curl --location '127.0.0.1:3000/jwt/login'

6. Visit 127.0.0.1:3000/jwt to check token
curl --location '127.0.0.1:3000/jwt' \
--header 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXkiOiIxIiwic3ViIjoidXNlciIsImV4cCI6MTY3NzU5OTIzMiwiaWF0IjoxNjc3NTk1NjMyfQ.3NY3SNvFE_2vHJAuBH1QwhPyTA_CtiV8y4w8nC1J5eM'
 ********************************/

type JwtController struct {
	// Dependent services
}

func NewJwtController() *JwtController {
	return &JwtController{
		// Inject services
	}
}

func (r *JwtController) Login(ctx http.Context) http.Response {
	user := &models.User{
		Name:   "Goravel",
		Avatar: "logo.png",
	}
	if err := facades.Orm().Query().Create(&user); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	token, err := facades.Auth(ctx).LoginUsingID(user.ID)
	if err != nil {
		return ctx.Response().String(http.StatusInternalServerError, err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"token": token,
	})
}

func (r *JwtController) Index(ctx http.Context) http.Response {
	return ctx.Response().Success().Json(http.Json{
		"token": ctx.Request().Header("Authorization", ""),
	})
}
