package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type InertiaController struct{}

func NewInertiaController() *InertiaController {
	return &InertiaController{}
}

func (r *InertiaController) Home(ctx http.Context) http.Response {
	facades.Inertia().Defer(ctx, "stats", func() any { return map[string]any{"users": 1280} })
	return facades.Inertia().Render(ctx, "Home", map[string]any{"message": "Goravel + Inertia"})
}

func (r *InertiaController) Feed(ctx http.Context) http.Response {
	facades.Inertia().Merge(ctx, "items", func() any { return []map[string]any{{"id": 1, "title": "Item #1"}} })
	return facades.Inertia().Render(ctx, "Feed", map[string]any{"page": 1})
}

func (r *InertiaController) Contact(ctx http.Context) http.Response {
	return facades.Inertia().Render(ctx, "Contact", map[string]any{})
}

func (r *InertiaController) StoreContact(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{"email": "required|email"})
	if err != nil || validator.Fails() {
		if validator != nil {
			facades.Inertia().FlashErrors(ctx, validator.Errors())
		}
		return facades.Inertia().Redirect(ctx, "/inertia/contact")
	}
	ctx.Request().Session().Flash("success", "Saved!")
	return facades.Inertia().Redirect(ctx, "/inertia/contact")
}

func (r *InertiaController) RedirectTo(ctx http.Context) http.Response {
	return facades.Inertia().Redirect(ctx, "/inertia/home")
}

func (r *InertiaController) LocationTo(ctx http.Context) http.Response {
	return facades.Inertia().Location(ctx, "https://example.com")
}
