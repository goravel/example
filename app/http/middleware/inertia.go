package middleware

import (
	"github.com/goravel/framework/contracts/http"

	inertiamw "github.com/goravel/inertia/middleware"

	"goravel/app/facades"
)

func Inertia() http.Middleware {
	return inertiamw.Handle(inertiamw.Options{Share: share})
}

func share(ctx http.Context) map[string]any {
	return map[string]any{
		"appName": facades.Config().GetString("app.name"),
		"auth":    map[string]any{"user": authUser(ctx)},
	}
}

func authUser(ctx http.Context) any {
	if !ctx.Request().HasSession() {
		return nil
	}
	return ctx.Request().Session().Get("user")
}
