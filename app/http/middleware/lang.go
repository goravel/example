package middleware

import (
	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

func Lang() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		lang := ctx.Request().Input("lang")
		if lang == "" {
			lang = facades.Config().GetString("app.locale")
		}
		facades.App().SetLocale(ctx, lang)

		ctx.Request().Next()
	}
}
