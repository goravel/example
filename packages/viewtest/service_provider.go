package viewtest

import (
	"html/template"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/path"

	"goravel/app/facades"
)

// GreetingFuncName is registered by both this provider and the application, so
// a rename breaks the build instead of quietly dropping the override coverage.
const GreetingFuncName = "greeting"

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

	// AddFuncs runs in provider Boot; the application's WithViewFuncs is applied
	// after every provider has booted, so the app wins this name.
	greet := func(name string) string {
		return "package-hello " + name
	}
	facades.View().AddFuncs(template.FuncMap{
		"package_greeting": greet,
		GreetingFuncName:   greet,
	})
}
