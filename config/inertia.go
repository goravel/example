package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	// The demo pins an asset version so the inertia tests and their clients agree;
	// leave the fallback below empty to follow the Vite build manifest instead.
	config.Add("inertia", map[string]any{
		"root_view":   "resources/inertia/app.gohtml",
		"version":     config.Env("INERTIA_VERSION", "test-version"),
		"ssr":         config.Env("INERTIA_SSR", false),
		"ssr_url":     config.Env("INERTIA_SSR_URL", "http://127.0.0.1:13714/render"),
		"ssr_timeout": config.Env("INERTIA_SSR_TIMEOUT", 5),
		"flash_keys":  []string{"success", "error", "warning", "info", "message"},
	})
}
