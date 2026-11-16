package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("modular", map[string]any{
		// Directory of the application's own modules, relative to the project root.
		"path": "modules",
	})
}
