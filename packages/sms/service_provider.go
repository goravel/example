package sms

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "sms"

var App foundation.Application

type ServiceProvider struct {
}

// Relationship returns the relationship of the service provider.
func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

// Register registers the service provider.
func (r *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return nil, nil
	})
}

// Boot boots the service provider, will be called after all service providers are registered.
func (r *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("./packages/sms", map[string]string{
		"setup/config/sms.go": app.ConfigPath("sms.go"),
	})
}
