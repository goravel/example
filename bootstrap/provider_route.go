package bootstrap

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type RouteProvider struct{}

func (r *RouteProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

func (r *RouteProvider) Register(app foundation.Application) {}

func (r *RouteProvider) Boot(app foundation.Application) {
	facades.Route().Get("/provider-route", func(ctx http.Context) http.Response {
		return ctx.Response().Success().String("Hello from provider route")
	})
}
