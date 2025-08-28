package login

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type Captcha struct {
	ID   string `form:"id" json:"id"`
	Code string `form:"code" json:"code"`
}

type LoginRequest struct {
	Username string  `form:"username" json:"username"`
	Password string  `form:"password" json:"password"`
	Captcha  Captcha `form:"captcha" json:"captcha"`
	Origin   string  `form:"origin" json:"origin"`
}

func (r *LoginRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *LoginRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "trim",
		"origin":   "trim",
	}
}

func (r *LoginRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username":     "required|string|min:1|max:100",
		"password":     "required|string|min:5",
		"captcha.id":   "required|string",
		"captcha.code": "required|string",
		"origin":       "required|string|max:100",
	}
}

func (r *LoginRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required":     "用户名不能为空",
		"username.min":          "用户名长度不能少于1个字符",
		"username.max":          "用户名长度不能超过100个字符",
		"password.required":     "密码不能为空",
		"password.min":          "密码长度不能少于5个字符",
		"captcha.id.required":   "验证码ID不能为空",
		"captcha.code.required": "验证码不能为空",
		"origin.required":       "来源不能为空",
		"origin.max":            "来源信息长度不能超过100个字符",
	}
}

func (r *LoginRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"username":     "用户名",
		"password":     "密码",
		"captcha.id":   "验证码ID",
		"captcha.code": "验证码",
		"origin":       "来源",
	}
}

func (r *LoginRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
