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

type GateController struct {
	//Dependent services
}

func NewGateController() *GateController {
	return &GateController{
		//Inject services
	}
}

func (r *GateController) AllowCreate(ctx http.Context) {
	var user models.User
	user.ID = 1

	if facades.Gate.Allows("create-user", map[string]any{
		"user":   user,
		"userID": 1,
	}) {
		ctx.Response().Success().String("allow create user")
		return
	}

	ctx.Response().String(http.StatusMethodNotAllowed, "deny create user")
}

func (r *GateController) DenyUpdate(ctx http.Context) {
	var user models.User
	user.ID = 1

	if facades.Gate.Allows("update-user", map[string]any{
		"user":   user,
		"userID": 2,
	}) {
		ctx.Response().Success().String("allow create user")
		return
	}

	ctx.Response().String(http.StatusMethodNotAllowed, "deny create user")
}

func (r *GateController) Inspect(ctx http.Context) {
	var user models.User
	user.ID = 1

	response := facades.Gate.Inspect("create-user", map[string]any{
		"user":   user,
		"userID": 1,
	})

	if response.Allowed() {
		ctx.Response().Success().String("allow create user")
		return
	}

	ctx.Response().String(http.StatusMethodNotAllowed, response.Message())
}

func (r *GateController) Before(ctx http.Context) {
	var user models.User
	user.ID = 1

	response := facades.Gate.Inspect("create-user", map[string]any{
		"user":   user,
		"userID": 1,
	})

	if response.Allowed() {
		ctx.Response().Success().String("allow create user")
		return
	}

	ctx.Response().String(http.StatusMethodNotAllowed, response.Message())
}

func (r *GateController) After(ctx http.Context) {
	var user models.User
	user.ID = 1

	response := facades.Gate.Inspect("create-user", map[string]any{
		"user":   user,
		"userID": 1,
	})

	if response.Allowed() {
		ctx.Response().Success().String("allow create user")
		return
	}

	ctx.Response().String(http.StatusMethodNotAllowed, response.Message())
}
