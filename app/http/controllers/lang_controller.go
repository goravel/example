package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type LangController struct {
	// Dependent services
}

func NewLangController() *LangController {
	return &LangController{
		// Inject services
	}
}

func (r *LangController) Index(ctx http.Context) http.Response {
	return ctx.Response().Success().Json(http.Json{
		"current_locale": facades.App().CurrentLocale(ctx),
		"name":           facades.Lang(ctx).Get("name"),
		"fallback":       facades.Lang(ctx).Get("description"),
	})
}
