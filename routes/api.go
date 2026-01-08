package routes

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	httpmiddleware "github.com/goravel/framework/http/middleware"

	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Api() {
	// Auth
	authController := controllers.NewAuthController()
	facades.Route().Prefix("jwt").Group(func(route route.Router) {
		route.Post("login", authController.LoginByJwt)
		route.Middleware(middleware.Jwt()).Get("info", authController.InfoByJwt)
	})

	facades.Route().Prefix("session").Group(func(route route.Router) {
		route.Post("login", authController.LoginBySession)
		route.Middleware(middleware.Session()).Get("info", authController.InfoBySession)
	})

	// DB
	dbController := controllers.NewDBController()
	facades.Route().Get("/db", dbController.Index)

	// Websocket
	websocketController := controllers.NewWebsocketController()
	facades.Route().Get("/ws", websocketController.Server)

	// Validation
	validationController := controllers.NewValidationController()
	facades.Route().Post("/validation/json", validationController.Json)
	facades.Route().Post("/validation/request", validationController.Request)
	facades.Route().Post("/validation/form", validationController.Form)

	// Localization
	langController := controllers.NewLangController()
	facades.Route().Middleware(middleware.Lang()).Get("lang", langController.Index)

	// Test Rate Limiter
	facades.Route().Middleware(httpmiddleware.Throttle("ip")).Get("/throttle", func(ctx http.Context) http.Response {
		return ctx.Response().Success().String("success")
	})

	facades.Route().Middleware(middleware.Localization()).Resource("users", controllers.NewUserController())

	facades.Route().Get("stream", func(ctx http.Context) http.Response {
		return ctx.Response().Stream(http.StatusCreated, func(w http.StreamWriter) error {
			data := []string{"a", "b", "c"}
			for _, item := range data {
				if _, err := w.Write([]byte(item + "\n")); err != nil {
					return err
				}

				if err := w.Flush(); err != nil {
					return err
				}

				time.Sleep(1 * time.Second)
			}

			return nil
		})
	})

	facades.Route().Get("timeout", func(ctx http.Context) http.Response {
		time.Sleep(10 * time.Second)
		return ctx.Response().String(http.StatusOK, "ok")
	})

	facades.Route().Get("panic", func(ctx http.Context) http.Response {
		panic("test panic")
	})

	facades.Route().Get("bind-query", func(ctx http.Context) http.Response {
		type Query struct {
			Name string `form:"name"`
		}

		var query Query
		if err := ctx.Request().BindQuery(&query); err != nil {
			return ctx.Response().Json(http.StatusBadRequest, http.Json{
				"error": err.Error(),
			})
		}

		return ctx.Response().Json(http.StatusOK, http.Json{
			"name": query.Name,
		})
	})

	facades.Route().Post("input-map", func(ctx http.Context) http.Response {
		return ctx.Response().Json(http.StatusOK, http.Json{
			"test": ctx.Request().InputMap("test"),
		})
	})

	facades.Route().Post("input-map-array", func(ctx http.Context) http.Response {
		return ctx.Response().Json(http.StatusOK, http.Json{
			"test": ctx.Request().InputMapArray("test"),
		})
	})

	facades.Route().Post("files", func(ctx http.Context) http.Response {
		files, err := ctx.Request().Files("files")
		if err != nil {
			return ctx.Response().Json(http.StatusBadRequest, http.Json{
				"error": err.Error(),
			})
		}

		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.GetClientOriginalName())
		}

		return ctx.Response().Json(http.StatusOK, http.Json{
			"files": fileNames,
		})
	})

	facades.Route().Prefix("url").Group(func(route route.Router) {
		route.Get("get/{id}", func(ctx http.Context) http.Response {
			return ctx.Response().Json(http.StatusOK, http.Json{
				"full_url":    ctx.Request().FullUrl(),
				"info":        ctx.Request().Info(),
				"info1":       facades.Route().Info("url.get"),
				"method":      ctx.Request().Method(),
				"name":        ctx.Request().Name(),
				"origin_path": ctx.Request().OriginPath(),
				"path":        ctx.Request().Path(),
				"url":         ctx.Request().Url(),
			})
		}).Name("url.get")
		route.Post("post/{id}", func(ctx http.Context) http.Response {
			return ctx.Response().Json(http.StatusOK, http.Json{
				"full_url":    ctx.Request().FullUrl(),
				"info":        ctx.Request().Info(),
				"info1":       facades.Route().Info("url.post"),
				"method":      ctx.Request().Method(),
				"name":        ctx.Request().Name(),
				"origin_path": ctx.Request().OriginPath(),
				"path":        ctx.Request().Path(),
				"url":         ctx.Request().Url(),
			})
		}).Name("url.post")
	})
}
