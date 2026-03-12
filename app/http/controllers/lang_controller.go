package controllers

import (
	"sync"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
)

type LangController struct {
	// Dependent services
}

func NewLangController() *LangController {
	return &LangController{
		// Inject services
	}
}

// Index lang index
// @Summary lang index
// @Description lang index
// @Tags Lang
// @Accept json
// @Success 200 {object} map[string]any
// @Router /lang [get]
func (r *LangController) Index(ctx http.Context) http.Response {
	return ctx.Response().Success().Json(http.Json{
		"current_locale": facades.App().CurrentLocale(ctx),
		"name":           facades.Lang(ctx).Get("name"),
		"fallback":       facades.Lang(ctx).Get("description"),
		"fs":             facades.Lang(ctx).Get("fs"),
	})
}

var (
	LangControllerSingleton *LangController
	langControllerOnce      sync.Once
)

func (r *LangController) Singleton() *LangController {

	langControllerOnce.Do(func() {
		LangControllerSingleton = NewLangController()
	})

	return LangControllerSingleton
}

// Routes Lang routes.
// Example Usage:
// @api|web.go: controllers.LangControllerSingleton.Routes(nil)
func (r *LangController) Routes(baseRouter route.Router) {
	r.Singleton()
	var LangRouter = baseRouter
	if LangRouter == nil {
		LangRouter = facades.Route()
	}
	LangRouter.
		Get(
			"/lang",

			LangControllerSingleton.
				Index)

}
