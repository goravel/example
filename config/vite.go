package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("vite", map[string]any{
		// Build Directory
		//
		// The directory inside the public directory that `vite build` writes to.
		"build_directory": "build",

		// Hot File
		//
		// The file the dev server announces its URL in, relative to the project
		// root. While it exists, {{ vite }} points every tag at the dev server.
		"hot_file": "public/hot",

		// Manifest
		//
		// The manifest `vite build` writes, relative to the build directory.
		"manifest": ".vite/manifest.json",
	})
}
