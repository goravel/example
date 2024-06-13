package middleware

import (
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func Lang() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		facades.App().SetLocale(ctx, ctx.Request().Input("lang", facades.Config().GetString("app.locale")))

		ctx.Request().Next()
	}
}
