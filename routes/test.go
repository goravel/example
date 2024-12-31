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
}
