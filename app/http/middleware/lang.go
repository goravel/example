package middleware

import (
	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type langMiddleware struct{}

func (r *langMiddleware) Signature() string {
	return "lang"
}

func (r *langMiddleware) Handle(ctx httpcontract.Context) {
	lang := ctx.Request().Input("lang")
	if lang == "" {
		lang = facades.Config().GetString("app.locale")
	}
	facades.App().SetLocale(ctx, lang)

	ctx.Request().Next()
}

func Lang() httpcontract.Middleware {
	return &langMiddleware{}
}
