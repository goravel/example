package middleware

import (
	"goravel/app/facades"
	"log"

	"github.com/goravel/framework/contracts/http"
)

func AuthUser() http.Middleware {
	return func(ctx http.Context) {

		log.Println("AuthUser middleware running")

		token := ctx.Request().Cookie("token", "")
		if token == "" {
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/user/signIn")
			ctx.Request().Abort()
			return
		}

		payload, err := facades.Auth(ctx).Parse(token)
		if err != nil {
			facades.Log().Error("Token invalid:", err)
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/user/signIn")
			ctx.Request().Abort()
			return
		}

		if payload.Guard != "users" {
			facades.Log().Warning("Token bukan user")
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/user/signIn")
			ctx.Request().Abort()
			return
		}

		if facades.Auth(ctx).Guard("admins").Check() {
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/admin/")
			ctx.Request().Abort()
			return
		}

		if !facades.Auth(ctx).Guard("users").Check() {
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/user/signIn")
			ctx.Request().Abort()
			return
		}

		ctx.Request().Next()
	}
}
