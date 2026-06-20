package config

import (
	"github.com/goravel/framework/support/carbon"

	"goravel/app/facades"
	"goravel/lang_fs"
)

// Boot Start all init methods of the current folder to bootstrap all config.
func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		// Application Name
		//
		// This value is the name of your application. This value is used when the
		// framework needs to place the application's name in a notification or
		// any other location as required by the application or its packages.
		"name": config.Env("APP_NAME", "Goravel"),

		// Application Environment
		//
		// This value determines the "environment" your application is currently
		// running in. This may determine how you prefer to configure various
		// services the application utilizes. Set this in your ".env" file.
		"env": config.Env("APP_ENV", "production"),

		// Application Debug Mode
		"debug": config.Env("APP_DEBUG", false),

		// Application Timezone
		//
		// Here you may specify the default timezone for your application.
		// Example: UTC, Asia/Shanghai
		// More: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
		"timezone": carbon.UTC,

		// Application Locale Configuration
		//
		// The application locale determines the default locale that will be used
		// by the translation service provider. You are free to set this value
		// to any of the locales which will be supported by the application.
		"locale": "en",

		// Application Fallback Locale
		//
		// The fallback locale determines the locale to use when the current one
		// is not available. You may change the value to correspond to any of
		// the language folders that are provided through your application.
		"fallback_locale": "cn",

		// Application Lang Path
		//
		// The path to the language files for the application. You may change
		// the path to a different directory if you would like to customize it.
		"lang_path": "lang",
		"lang_fs":   lang_fs.Fs,

		// Encryption Key
		//
		// 32 character string, otherwise these encrypted strings
		// will not be safe. Please do this before deploying an application!
		"key": config.Env("APP_KEY", ""),

		// Maintenance Mode
		//
		// This value determines the driver used to store the maintenance mode
		// state. Supported drivers: "file", "cache". If you use the "cache"
		// driver, you may specify a cache store name.
		"maintenance": map[string]any{
			"driver": config.Env("APP_MAINTENANCE_DRIVER", "file"),
			"store":  config.Env("APP_MAINTENANCE_STORE", ""),
		},

		// Disabled Runners
		//
		// Here you may specify which auto-run runners should be skipped when
		// app.Start() is called. Patterns use path.Match glob matching.
		//
		// Framework runner signatures:
		//   goravel:http      – HTTP server
		//   goravel:grpc      – gRPC server
		//   goravel:queue     – Queue worker
		//   goravel:schedule  – Scheduler
		//   goravel:telemetry – Telemetry
		//
		// Example: disable scheduler for a web-only container
		//   "disabled_runners": []string{"goravel:schedule"},
		//
		// Example: disable all framework runners
		//   "disabled_runners": []string{"goravel:*"},
		"disabled_runners": []string{},
	})
}
