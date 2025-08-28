package controllers

import (
	"goravel/app/http/requests/login"

	"github.com/goravel/framework/contracts/http"
)

type LoginController struct {
}

func NewLoginController() *LoginController {
	return &LoginController{}
}

// Login 用户登录
func (c *LoginController) Login(ctx http.Context) http.Response {
	var loginRequest login.LoginRequest

	errors, err := ctx.Request().ValidateRequest(&loginRequest)
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"code":    1,
			"type":    "validation1",
			"message": err.Error(),
		})
	}

	if errors != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"code":    1,
			"type":    "validation2",
			"message": errors.All(),
		})
	}

	// TODO: 实现具体的登录逻辑
	return ctx.Response().Json(http.StatusOK, http.Json{
		"code":    0,
		"message": "登录成功",
		"data": http.Json{
			"username": loginRequest.Username,
			"origin":   loginRequest.Origin,
		},
	})
}
