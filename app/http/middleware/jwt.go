package middleware

import (
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/facades"
)

func Jwt() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			ctx.Request().AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if _, err := facades.Auth(ctx).Parse(token); err != nil {
			if errors.Is(err, errors.AuthTokenExpired) {
				token, err = facades.Auth(ctx).Refresh()
				if err != nil {
					// Refresh time exceeded
					ctx.Request().AbortWithStatus(http.StatusUnauthorized)
					return
				}

				token = "Bearer " + token
			} else {
				ctx.Request().AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		// You can get User in DB and set it to ctx

		//var user models.User
		//if err := facades.Auth().User(ctx, &user); err != nil {
		//	ctx.Request().AbortWithStatus(http.StatusUnauthorized)
		//  return
		//}
		//ctx.WithValue("user", user)

		ctx.Response().Header("Authorization", token)
		ctx.Request().Next()
	}
}
