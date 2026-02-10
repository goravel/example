package controllers

import (
	// "github.com/bgdar/myBakery/app/facades"
	// "github.com/bgdar/myBakery/app/models"
	"fmt"
	"goravel/app/facades"
	"goravel/app/models"
	"log"
	"time"

	"github.com/goravel/framework/contracts/http"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (r *UserController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("user/index.html", map[string]any{
		"title": "User index ",
	})
}

func (r *UserController) Profile(ctx http.Context) http.Response {
	return ctx.Response().View().Make("user/profile.html", map[string]any{
		"title": "User Page",
	})
}

func (r *UserController) SignInView(ctx http.Context) http.Response {

	return ctx.Response().View().Make("user/signIn.html", map[string]any{
		"title": "Home | My Bakery",
		// "flash": ctx.Request().Session().Get("flash"), // ambil flash jika mau render jadi gak perlu refres halaman
	})

}

func (r *UserController) SignInPost(ctx http.Context) http.Response {

	log.Println("Route userssignIN post ")
	email := ctx.Request().Origin().FormValue("email")
	password := ctx.Request().Origin().FormValue("password")

	var user models.Users

	err := facades.Orm().
		Query().
		Table("users").
		Where("email = ?", email).
		First(&user)

	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "gagal mendapatkan data user",
		})
	}

	if !facades.Hash().Check(password, user.Password) {
		log.Println("password salah")
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "password salah ",
		})
	}

	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "gagall mendpatkan token",
		})
	}

	ctx.Response().Cookie(http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   false,
		HttpOnly: true,
	})
	return ctx.Response().Redirect(http.StatusTemporaryRedirect, "/user/profile")
}

func (r *UserController) SignUpView(ctx http.Context) http.Response {

	return ctx.Response().View().Make("user/signUp.html", map[string]any{
		"title": "Home | My Bakery",
	})

}

func (r *UserController) SignUpPost(ctx http.Context) http.Response {

	log.Println("Route user sinUp post ")
	name := ctx.Request().Origin().FormValue("name")
	email := ctx.Request().Origin().FormValue("email")
	password := ctx.Request().Origin().FormValue("password")
	code_rendem := ctx.Request().Origin().FormValue("code_rendem")

	isUser, err := facades.Orm().Query().
		Table("users").
		Where("email", email).OrWhere("name", name).
		Exists()
	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "gagal mendapatkan data user ",
		})
	}
	log.Println("code rendem", code_rendem)

	if isUser {
		facades.Log().Info("User sudah ada")
		ctx.Request().Session().Flash("error", "User sudah terdaftar")
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "user sudah terdaftar",
		})
	}

	hashedPassword, err := facades.Hash().Make(password)
	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "gagal hashing password",
		})
	}
	var admin models.Admin
	var err_admin = facades.DB().Table("admins").Where("code_rendem", code_rendem).First(&admin)
	if err_admin != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "code rendem tidak di dapatkan , chat yout owner",
		})
	}

	log.Println("admin id")

	if err := facades.Orm().Query().Table("users").Create(map[string]any{
		"name":     name,
		"email":    email,
		"password": hashedPassword,
		"admin_id": admin.ID,
	}); err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": fmt.Sprintf("gagal menyimpan user : %s", err),
		})
	}

	return ctx.Response().Redirect(http.StatusTemporaryRedirect, "/user/signIn/")
}
