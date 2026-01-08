// app/http/middleware/Localization.go
package middleware

import (
	"fmt"

	"github.com/goravel/framework/contracts/http"
)

func Localization() http.Middleware {
	return func(ctx http.Context) {
		fmt.Println("Localization 开始")

		ctx.Request().Next()
		fmt.Println("Localization 结束")
	}

}
