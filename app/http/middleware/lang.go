package middleware

import (
	httpcontract "github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

func Lang() httpcontract.Middleware {
	return &LangMiddleware{}
}

type LangMiddleware struct{}

func (l *LangMiddleware) Handle(ctx httpcontract.Context) {
	lang := ctx.Request().Input("lang")
	if lang == "" {
		lang = facades.Config().GetString("app.locale")
	}
	facades.App().SetLocale(ctx, lang)

	ctx.Request().Next()
}

func (l *LangMiddleware) Signature() string {
	return "lang"
}
