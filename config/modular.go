package config

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("modular", map[string]any{
		// Directory of the application's own modules, relative to the project root.
		"path": "modules",

		// Route groups wrapping what modules register with Routes() and ApiRoutes().
		"routes": map[string]any{
			"web": map[string]any{"prefix": "", "middleware": []http.Middleware{}},
			"api": map[string]any{"prefix": "api", "middleware": []http.Middleware{}},
		},
	})
}
