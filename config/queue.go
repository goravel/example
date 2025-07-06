package config

import (
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	redisfacades "github.com/goravel/redis/facades"
)

func init() {
	config := facades.Config()
	config.Add("queue", map[string]any{
		// Default Queue Connection Name
		"default": config.Env("QUEUE_CONNECTION", "database"),

		// Queue Connections
		//
		// Here you may configure the connection information for each server that is used by your application.
		// Drivers: "sync", "database", "custom"
		"connections": map[string]any{
			"sync": map[string]any{
				"driver": "sync",
			},
			"database": map[string]any{
				"driver":     "database",
				"connection": "sqlite",
				"queue":      "default",
				"concurrent": 1,
			},
			"machinery": map[string]any{
				"driver":     "machinery",
				"connection": "default",
				"queue":      "default",
				"concurrent": 1,
			},
			"redis1": map[string]any{
				"driver":     "custom",
				"connection": "default",
				"queue":      "default",
				"via": func() (queue.Driver, error) {
					return redisfacades.Queue("redis1") // The `redis` value is the key of `connections`
				},
			},
			"redis": map[string]any{
				"driver":     "custom",
				"connection": "default",
				"queue":      "default",
				"via": func() (queue.Driver, error) {
					return redisfacades.Queue("redis") // The `redis` value is the key of `connections`
				},
			},
		},

		// Failed Queue Jobs
		//
		// These options configure the behavior of failed queue job logging so you
		// can control how and where failed jobs are stored.
		"failed": map[string]any{
			"database": config.Env("DB_CONNECTION", "postgres"),
			"table":    "failed_jobs",
		},
	})
}
