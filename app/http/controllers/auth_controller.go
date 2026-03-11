package controllers

import (
	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"

	"goravel/app/facades"
	"goravel/app/models"
)

/*********************************
Introduce JWT and Session auth

JWT:

1. Generate JWT secret
go run . artisan jwt:secret

2. Generate Middleware
go run . artisan make:middleware Jwt

3. Add route to `/route/api.go`

4. Run Server
air

5. Visit 127.0.0.1:3000/jwt/login to get token
curl -X POST -i http://127.0.0.1:3000/jwt/login

6. Visit 127.0.0.1:3000/jwt/info to check token
curl -X GET -i http://127.0.0.1:3000/jwt/info \
-H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXkiOiIxIiwic3ViIjoidXNlciIsImV4cCI6MTY3NzU5OTIzMiwiaWF0IjoxNjc3NTk1NjMyfQ.3NY3SNvFE_2vHJAuBH1QwhPyTA_CtiV8y4w8nC1J5eM'

Session:

1. Generate Middleware
go run . artisan make:middleware Session

2. Add route to `/route/api.go`

3. Run Server
air

4. Visit 127.0.0.1:3000/session/login to get token
curl -X POST -i http://127.0.0.1:3000/session/login

5. Visit 127.0.0.1:3000/session/info to check token
curl -X GET -i http://127.0.0.1:3000/session/info \
-H 'Guard: session' \
-b 'goravel_session=zI2I5E6BOa5ojT8CVcxf8t0SUzct2kOV2BtklnHv; Path=/; Max-Age=7199; HttpOnly; SameSite=Lax'

 ********************************/

type AuthController struct {
	// Dependent services
}

func NewAuthController() *AuthController {
	return &AuthController{
		// Inject services
	}
}

func (r *AuthController) authDriver(ctx http.Context) (string, contractsauth.GuardDriver) {
	guard := ctx.Request().Header("Guard")
	if guard == "" {
		guard = facades.Config().GetString("auth.defaults.guard")
	}

	return guard, facades.Auth(ctx).Guard(guard)
}

func (r *AuthController) LoginByJwt(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().FirstOrCreate(&user, models.User{
		Name: ctx.Request().Input("name", "Goravel"),
	}); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	var (
		token string
		err   error
	)

	// Use different guards to login
	if guard := ctx.Request().Header("Guard"); guard == "" {
		token, err = facades.Auth(ctx).LoginUsingID(user.ID)
		if err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
	} else {
		token, err = facades.Auth(ctx).Guard(guard).Login(user)
		if err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
	}

	return ctx.Response().Header("Authorization", "Bearer "+token).Success().Json(http.Json{
		"user": user,
	})
}

func (r *AuthController) LoginUsingID(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().FirstOrCreate(&user, models.User{
		Name: ctx.Request().Input("name", "Goravel"),
	}); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	guard, auth := r.authDriver(ctx)
	token, err := auth.LoginUsingID(user.ID)
	if err != nil {
		return ctx.Response().String(http.StatusInternalServerError, err.Error())
	}

	if token != "" {
		ctx.Response().Header("Authorization", "Bearer "+token)
	}

	return ctx.Response().Success().Json(http.Json{
		"guard": guard,
		"token": token,
		"user":  user,
	})
}

func (r *AuthController) InfoByJwt(ctx http.Context) http.Response {
	var (
		id   string
		user models.User
		err  error
	)

	if guard := ctx.Request().Header("Guard"); guard == "" {
		if err := facades.Auth(ctx).User(&user); err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
		id, err = facades.Auth(ctx).ID()

	} else {
		if err := facades.Auth(ctx).Guard(guard).User(&user); err != nil {
			return ctx.Response().String(http.StatusInternalServerError, err.Error())
		}
		id, err = facades.Auth(ctx).Guard(guard).ID()
	}

	if err != nil {
		return ctx.Response().String(http.StatusInternalServerError, err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"id":   cast.ToUint(id),
		"user": user,
	})
}

func (r *AuthController) State(ctx http.Context) http.Response {
	guard, auth := r.authDriver(ctx)

	id, idErr := auth.ID()

	var user models.User
	userErr := auth.User(&user)

	return ctx.Response().Success().Json(http.Json{
		"check":      auth.Check(),
		"guard":      guard,
		"guest":      auth.Guest(),
		"id":         id,
		"id_error":   idErr != nil,
		"user":       user,
		"user_error": userErr != nil,
	})
}

func (r *AuthController) Parse(ctx http.Context) http.Response {
	_, auth := r.authDriver(ctx)

	payload, err := auth.Parse(ctx.Request().Header("Authorization"))

	response := http.Json{
		"error": err != nil,
	}
	if payload != nil {
		response["guard"] = payload.Guard
		response["key"] = payload.Key
	}

	return ctx.Response().Success().Json(response)
}

func (r *AuthController) Refresh(ctx http.Context) http.Response {
	_, auth := r.authDriver(ctx)

	token, err := auth.Refresh()

	response := http.Json{
		"error": err != nil,
		"token": token,
	}

	if token != "" {
		ctx.Response().Header("Authorization", "Bearer "+token)
	}

	return ctx.Response().Success().Json(response)
}

func (r *AuthController) Logout(ctx http.Context) http.Response {
	_, auth := r.authDriver(ctx)
	if err := auth.Logout(); err != nil {
		return ctx.Response().String(http.StatusInternalServerError, err.Error())
	}

	return ctx.Response().Success().Json(http.Json{
		"logged_out": true,
	})
}

func (r *AuthController) LoginBySession(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().FirstOrCreate(&user, models.User{
		Name: ctx.Request().Input("name", "Goravel"),
	}); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	if _, err := facades.Auth(ctx).Guard("session").Login(user); err != nil {
		return ctx.Response().String(http.StatusInternalServerError, err.Error())
	}

	return ctx.Response().Header("Guard", "session").Success().Json(http.Json{
		"user": user,
	})
}

func (r *AuthController) InfoBySession(ctx http.Context) http.Response {
	user := ctx.Value("user").(models.User)

	return ctx.Response().Success().Json(http.Json{
		"id":   user.ID,
		"user": user,
	})
}
