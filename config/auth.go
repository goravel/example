package config

import (
	"goravel/app/facades"
	"goravel/app/models"
)

func init() {
	config := facades.Config()
	config.Add("auth", map[string]any{
		// Authentication Defaults
		//
		// This option controls the default authentication "guard"
		// reset options for your application. You may change these defaults
		// as required, but they're a perfect start for most applications.
		"defaults": map[string]any{
			"guard": "user",
		},

		// Authentication Guards
		//
		// Next, you may define every authentication guard for your application.
		// Of course, a great default configuration has been defined for you
		// here which uses session storage and the Eloquent user provider.
		//
		// All authentication drivers have a user provider. This defines how the
		// users are actually retrieved out of your database or other storage
		// mechanisms used by this application to persist your user's data.
		//
		// Supported drivers: "jwt", "session"
		"guards": map[string]any{
			"users": map[string]any{
				"driver":      "jwt",
				"provider":    "users",
				"ttl":         60,
				"refresh_ttl": 0,
				// "secret":      facades.Config().Env("JWT_SECRET", " "),
			},
			// add admin guard
			"admins": map[string]any{
				"driver":      "jwt",
				"provider":    "admins",
				"ttl":         60,
				"refresh_ttl": 0,
				// "secret":      facades.Config().Env("JWT_SECRET", " "),
			},
		},

		// Supported: "orm"
		"providers": map[string]any{
			"users": map[string]any{
				"driver": "orm",
				"model":  &models.Users{},
			},
			"admins": map[string]any{
				"driver": "orm",
				"model":  &models.Admin{},
			},
		},
	})
}
