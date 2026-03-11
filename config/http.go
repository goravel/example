package config

import (
	"github.com/gin-gonic/gin/render"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	fiberfacades "github.com/goravel/fiber/facades"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/gin"
	ginfacades "github.com/goravel/gin/facades"

	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("http", map[string]any{
		// HTTP Driver
		"default": "gin",
		// HTTP Drivers
		"drivers": map[string]any{
			"gin": map[string]any{
				// Optional, default is 4096 KB
				"body_limit":   4096,
				"header_limit": 4096,
				"route": func() (route.Route, error) {
					return ginfacades.Route("gin"), nil
				},
				// Optional, default is http/template
				"template": func() (render.HTMLRender, error) {
					return gin.DefaultTemplate()
				},
			},
			"fiber": map[string]any{
				// immutable mode, see https://docs.gofiber.io/#zero-allocation
				// WARNING: This option is dangerous. Only change it if you fully understand the potential consequences.
				"immutable": true,
				// prefork mode, see https://docs.gofiber.io/api/fiber/#config
				"prefork": false,
				// Optional, default is 4096 KB
				"body_limit":   4096,
				"header_limit": 4096,
				"route": func() (route.Route, error) {
					return fiberfacades.Route("fiber"), nil
				},
				// Optional, default is "html/template"
				"template": func() (fiber.Views, error) {
					return html.New(path.Resource("views"), ".tmpl"), nil
				},
			},
		},
		// HTTP URL
		"url": config.Env("APP_URL", "http://localhost"),
		// HTTP Host
		"host": config.Env("APP_HOST", "127.0.0.1"),
		// HTTP Port
		"port": config.Env("APP_PORT", "3000"),
		// HTTP Timeout, default is 3 seconds
		"request_timeout": 3,
		// HTTPS Configuration
		"tls": map[string]any{
			// HTTPS Host
			"host": config.Env("APP_HOST", "127.0.0.1"),
			// HTTPS Port
			"port": config.Env("APP_PORT", "3000"),
			// SSL Certificate, you can put the certificate in /public folder
			"ssl": map[string]any{
				// ca.pem
				"cert": "",
				// ca.key
				"key": "",
			},
		},
		"default_client": config.Env("HTTP_CLIENT_DEFAULT", "default"),
		"clients": map[string]any{
			"default": map[string]any{
				"base_url":                config.Env("HTTP_CLIENT_BASE_URL", "http://127.0.0.1:3000"),
				"timeout":                 config.Env("HTTP_CLIENT_TIMEOUT", "30s"),
				"max_idle_conns":          config.Env("HTTP_CLIENT_MAX_IDLE_CONNS", 100),
				"max_idle_conns_per_host": config.Env("HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST", 2),
				"max_conns_per_host":      config.Env("HTTP_CLIENT_MAX_CONN_PER_HOST", 0),
				"idle_conn_timeout":       config.Env("HTTP_CLIENT_IDLE_CONN_TIMEOUT", "90s"),
			},
		},
	})
}
