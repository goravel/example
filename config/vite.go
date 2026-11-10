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
		// The file the dev server announces its URL in, inside the public
		// directory. While it exists, {{ vite }} points every tag at the dev server.
		"hot_file": "hot",

		// Manifest
		//
		// The manifest `vite build` writes, relative to the build directory.
		"manifest": ".vite/manifest.json",

		// Inputs and Assets
		//
		// Entries and assets `vite build` must include even when no view names
		// them; vite:inputs merges these with the view calls, once each and
		// sorted. The views already name both below, so the build is unchanged.
		"inputs": []string{"resources/inertia/app.ts"},
		"assets": []string{"resources/images/logo.svg"},
	})
}
