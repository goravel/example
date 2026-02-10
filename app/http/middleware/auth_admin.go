package middleware

import (
	"goravel/app/facades"
	"log"

	"github.com/goravel/framework/contracts/http"
)

func AuthAdmin() http.Middleware {
	return func(ctx http.Context) {

		log.Println("AuthAdmin middleware running")

		token := ctx.Request().Cookie("token", "")
		if token == "" {
			ctx.Request().Abort()
			return
		}

		payload, err := facades.Auth(ctx).Parse(token)
		if err != nil {
			facades.Log().Error("Token invalid:", err)
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/admin/signIn")
			ctx.Request().Abort()
			return
		}

		if payload.Guard != "admins" {
			facades.Log().Warning("Token bukan admin")
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/admin/signIn")
			ctx.Request().Abort()
			return
		}

		if !facades.Auth(ctx).Guard("admins").Check() {
			ctx.Response().Redirect(http.StatusTemporaryRedirect, "/admin/signIn")
			ctx.Request().Abort()
			return
		}

		ctx.Request().Next()
	}
}
