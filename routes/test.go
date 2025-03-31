package routes

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/controllers"
)

func Test() {
	facades.Route().Resource("users", controllers.NewUserController())

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
}
