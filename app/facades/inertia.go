package facades

import (
	"github.com/goravel/inertia/contracts"
)

// Inertia returns the registered Inertia manager. The manager is bound as the
// "goravel.inertia" singleton by the Inertia ServiceProvider.
func Inertia() contracts.Inertia {
	instance, err := App().Make("goravel.inertia")
	if err != nil {
		panic(err)
	}

	return instance.(contracts.Inertia)
}
