package providers

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/services"
)

type TelemetryServiceProvider struct{}

// Register binds the example telemetry service as a singleton so its instruments
// are created once and reused. App().Restart() re-runs providers, so the
// singleton is rebuilt against the current telemetry provider.
func (r *TelemetryServiceProvider) Register(app foundation.Application) {
	app.Singleton(services.TelemetryBinding, func(app foundation.Application) (any, error) {
		return services.NewTelemetry()
	})
}

func (r *TelemetryServiceProvider) Boot(app foundation.Application) {}
