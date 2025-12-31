package middleware

import (
	"goravel/app/facades"

	httpcontract "github.com/goravel/framework/contracts/http"
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
