package controllers

import (
	frameworkauth "github.com/goravel/framework/auth"
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

func (r *AuthController) StatusByJwt(ctx http.Context) http.Response {
	guardDriver := r.jwtGuard(ctx)
	token := ctx.Request().Header("Authorization", "")

	response := http.Json{
		"check": guardDriver.Check(),
		"guest": guardDriver.Guest(),
	}

	if token == "" {
		return ctx.Response().Success().Json(response)
	}

	payload, err := guardDriver.Parse(token)
	if payload != nil {
		response["payload"] = http.Json{
			"guard": payload.Guard,
			"key":   payload.Key,
		}
	}
	if err != nil {
		response["parse_error"] = err.Error()
		return ctx.Response().Success().Json(response)
	}

	response["check"] = guardDriver.Check()
	response["guest"] = guardDriver.Guest()

	if id, err := guardDriver.ID(); err == nil {
		response["id"] = cast.ToUint(id)
	}

	var user models.User
	if err := guardDriver.User(&user); err == nil {
		response["user"] = user
	}

	return ctx.Response().Success().Json(response)
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

func (r *AuthController) RefreshByJwt(ctx http.Context) http.Response {
	guardDriver := r.jwtGuard(ctx)
	token := ctx.Request().Header("Authorization", "")

	if token == "" {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": "Authorization header is required",
		})
	}

	payload, err := guardDriver.Parse(token)
	response := http.Json{}
	if payload != nil {
		response["payload"] = http.Json{
			"guard": payload.Guard,
			"key":   payload.Key,
		}
	}
	if err != nil && err != frameworkauth.ErrorTokenExpired {
		response["parse_error"] = err.Error()
		return ctx.Response().Json(http.StatusUnauthorized, response)
	}
	if err != nil {
		response["parse_error"] = err.Error()
	}

	refreshedToken, err := guardDriver.Refresh()
	if err != nil {
		response["refresh_error"] = err.Error()
		return ctx.Response().Json(http.StatusUnauthorized, response)
	}

	return ctx.Response().Header("Authorization", "Bearer "+refreshedToken).Success().Json(response)
}

func (r *AuthController) LogoutByJwt(ctx http.Context) http.Response {
	guardDriver := r.jwtGuard(ctx)
	token := ctx.Request().Header("Authorization", "")

	if token == "" {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": "Authorization header is required",
		})
	}

	if _, err := guardDriver.Parse(token); err != nil {
		return ctx.Response().Json(http.StatusUnauthorized, http.Json{
			"error": err.Error(),
		})
	}

	if err := guardDriver.Logout(); err != nil {
		return ctx.Response().Json(http.StatusUnauthorized, http.Json{
			"error": err.Error(),
		})
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

func (r *AuthController) LoginBySessionUsingID(ctx http.Context) http.Response {
	id := ctx.Request().Input("id", "1")

	if _, err := facades.Auth(ctx).Guard("session").LoginUsingID(id); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	var user models.User
	if err := facades.Auth(ctx).Guard("session").User(&user); err != nil {
		return ctx.Response().Json(http.StatusUnauthorized, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Header("Guard", "session").Success().Json(http.Json{
		"id":   user.ID,
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

func (r *AuthController) StatusBySession(ctx http.Context) http.Response {
	guardDriver := facades.Auth(ctx).Guard("session")
	response := http.Json{
		"check": guardDriver.Check(),
		"guest": guardDriver.Guest(),
	}

	if id, err := guardDriver.ID(); err == nil {
		response["id"] = cast.ToUint(id)

		var user models.User
		if err := guardDriver.User(&user); err == nil {
			response["user"] = user
		}
	}

	return ctx.Response().Success().Json(response)
}

func (r *AuthController) LogoutBySession(ctx http.Context) http.Response {
	if err := facades.Auth(ctx).Guard("session").Logout(); err != nil {
		return ctx.Response().Json(http.StatusUnauthorized, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"logged_out": true,
	})
}

func (r *AuthController) UnsupportedBySession(ctx http.Context) http.Response {
	guardDriver := facades.Auth(ctx).Guard("session")
	_, parseErr := guardDriver.Parse(ctx.Request().Header("Authorization", ""))
	_, refreshErr := guardDriver.Refresh()

	return ctx.Response().Success().Json(http.Json{
		"parse_error":   parseErr.Error(),
		"refresh_error": refreshErr.Error(),
	})
}

func (r *AuthController) jwtGuard(ctx http.Context) contractsauth.GuardDriver {
	guard := ctx.Request().Header("Guard")
	if guard == "" {
		return facades.Auth(ctx)
	}

	return facades.Auth(ctx).Guard(guard)
}
