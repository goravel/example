package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("inertia", map[string]any{
		"root_view":   "resources/inertia/app.gohtml",
		"version":     config.Env("INERTIA_VERSION", "test-version"),
		"ssr":         config.Env("INERTIA_SSR", false),
		"ssr_url":     config.Env("INERTIA_SSR_URL", "http://127.0.0.1:13714/render"),
		"ssr_timeout": config.Env("INERTIA_SSR_TIMEOUT", 5),
		"flash_keys":  []string{"success", "error", "warning", "info", "message"},
		"vite": map[string]any{
			"public_path": "public",
			"build_dir":   "build",
			"hot_file":    "public/hot",
			"dev_url":     config.Env("VITE_DEV_URL", ""),
		},
	})
}
