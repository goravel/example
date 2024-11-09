package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/facades"
)

func Auth() http.Middleware {
	return func(ctx http.Context) {
		guard := facades.Config().GetString("auth.defaults.guard")
		if ctx.Request().Header("Guard") != "" {
			guard = ctx.Request().Header("Guard")
		}

		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			ctx.Request().AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if _, err := facades.Auth(ctx).Guard(guard).Parse(token); err != nil {
			if errors.Is(err, errors.AuthTokenExpired) {
				// Refresh token
				token, err = facades.Auth(ctx).Guard(guard).Refresh()
				if err != nil {
					// Refresh time exceeded
					ctx.Request().AbortWithStatus(http.StatusUnauthorized)
					return
				}

				token = "Bearer " + token
			} else {
				// Token is invalid
				ctx.Request().AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		// You can get User in DB and set it to ctx

		//var user models.User
		//if err := facades.Auth().User(ctx, &user); err != nil {
		//	ctx.Request().AbortWithStatus(http.StatusUnauthorized)
		//  return
		//}
		//ctx.WithValue("user", user)

		ctx.Response().Header("Authorization", token)
		ctx.Request().Next()
	}
}
