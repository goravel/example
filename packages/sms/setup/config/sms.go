package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("sms", map[string]any{
		"driver": "default",
	})
}
