package viewtest

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/path"

	"goravel/app/facades"
)

var App foundation.Application

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	App = app
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	facades.View().LoadViewsFrom(path.Base("packages", "viewtest", "views"))
}
