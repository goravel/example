package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("grpc", map[string]any{
		// Configure your server host
		"host": config.Env("GRPC_HOST", ""),

		// Configure your server port
		"port": config.Env("GRPC_PORT", ""),

		// Configure servers which the client will connect to
		"servers": map[string]any{
			//"user": map[string]any{
			//	"host":         config.Env("GRPC_USER_HOST", ""),
			//	"port":         config.Env("GRPC_USER_PORT", ""),
			//	"interceptors": []string{},
			//},
		},
	})
}
