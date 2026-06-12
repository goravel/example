package telemetry

import (
	"github.com/goravel/framework/facades"
)

// ConfigScope remembers the original values of overridden config keys so a
// suite can restore them in TearDownSuite.
type ConfigScope struct {
	saved map[string]any
}

// OverrideConfig snapshots the current value of every key in overrides,
// applies the overrides, and restarts the application so they take effect. On
// restart failure the returned scope is still usable to Restore.
func OverrideConfig(overrides map[string]any) (*ConfigScope, error) {
	saved := make(map[string]any, len(overrides))
	for key := range overrides {
		saved[key] = facades.Config().Get(key)
	}
	for key, value := range overrides {
		facades.Config().Add(key, value)
	}

	scope := &ConfigScope{saved: saved}
	if err := facades.App().Restart(); err != nil {
		return scope, err
	}

	return scope, nil
}

// Restore re-applies the snapshotted values and restarts the application.
func (r *ConfigScope) Restore() error {
	for key, value := range r.saved {
		facades.Config().Add(key, value)
	}

	return facades.App().Restart()
}
