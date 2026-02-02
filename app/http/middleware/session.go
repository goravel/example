package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
)

func Session() http.Middleware {
	return func(ctx http.Context) {
		guard := ctx.Request().Header("Guard")
		if guard == "" {
			ctx.Request().Abort(http.StatusUnauthorized)
			return
		}

		var user models.User
		if err := facades.Auth(ctx).Guard(guard).User(&user); err != nil {
			ctx.Request().Abort(http.StatusUnauthorized)
			return

		}

		if user.ID == 0 {
			ctx.Request().Abort(http.StatusUnauthorized)
			return
		}

		ctx.WithValue("user", user)
		ctx.Request().Next()
	}
}
