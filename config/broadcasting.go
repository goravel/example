package config

import "goravel/app/facades"

func init() {
	config := facades.Config()
	config.Add("broadcasting", map[string]any{
		"default": config.Env("BROADCAST_CONNECTION", "pusher"),

		"connections": map[string]any{
			"pusher": map[string]any{
				"driver": "pusher",
				"key":    config.Env("PUSHER_APP_KEY", "test-key"),
				"secret": config.Env("PUSHER_APP_SECRET", "test-secret"),
				"app_id": config.Env("PUSHER_APP_ID", "test-app"),
				"options": map[string]any{
					"cluster": config.Env("PUSHER_APP_CLUSTER", "mt1"),
					"host":    config.Env("PUSHER_HOST", "127.0.0.1"),
					"port":    config.Env("PUSHER_PORT", 6001),
					"scheme":  config.Env("PUSHER_SCHEME", "http"),
				},
			},
			"log": map[string]any{
				"driver": "log",
			},
			"null": map[string]any{
				"driver": "null",
			},
		},

		"auth": map[string]any{
			"enabled":    config.Env("BROADCAST_AUTH_ENABLED", true),
			"path":       config.Env("BROADCAST_AUTH_PATH", "/broadcasting/auth"),
			"middleware": []string{},
		},
	})
}
