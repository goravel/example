package controllers

import (
	"fmt"
	"goravel/app/facades"
	"goravel/app/models"
	"log"
	"time"

	"github.com/goravel/framework/contracts/http"
)

type AdminController struct {
	// Dependent services
}

func NewAdminController() *AdminController {
	return &AdminController{
		// Inject services
	}
}

func (r *AdminController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("admin/index.html")
}

func (r *AdminController) Profile(ctx http.Context) http.Response {
	return ctx.Response().View().Make("admin/profile.html")
}

func (r *AdminController) SignInView(ctx http.Context) http.Response {

	return ctx.Response().View().Make("admin/signIn.html", map[string]any{
		"title": "Home | My Bakery",
	})

}

func (r *AdminController) SignInPost(ctx http.Context) http.Response {

	email := ctx.Request().Origin().FormValue("email")
	password := ctx.Request().Origin().FormValue("password")

	var admin models.Admin

	log.Println("Data : ", email, password)
	log.Println("Data admin ", admin)

	query := facades.Orm().
		Query().
		Table("admins").
		Where("email", email).
		First(&admin)

	if query.Error() == "" {
		log.Println("Query error:", query)
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "admin tidak ditemukan",
		})
	}
	if !facades.Hash().Check(password, admin.Password) {
		log.Println("password salah")
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "password salah ",
		})
	}

	token, err := facades.Auth(ctx).Guard("admins").Login(&admin)
	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "gagall mendpatkan token",
		})
	}

	log.Println("Running 2")

	ctx.Response().Cookie(http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   false,
		HttpOnly: true,
	})
	return ctx.Response().Redirect(http.StatusTemporaryRedirect, "/admin/profile")
}

func (r *AdminController) SignUpView(ctx http.Context) http.Response {

	return ctx.Response().View().Make("admin/signUp.html", map[string]any{
		"title": "Home | My Bakery",
	})

}

func (r *AdminController) SignUpPost(ctx http.Context) http.Response {

	log.Println("sigUppost running ")
	var admin models.Admin
	err := ctx.Request().Bind(&admin)
	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "ada yang salah dengan field nya ",
		})
	}

	// Cek admin sudah ada
	isAdmin, err := facades.Orm().Query().
		Table("admins").
		Where("email", admin.Email).OrWhere("name", admin.Name).
		Exists()
	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "gagal mendapatkan data admin ",
		})
	}

	if isAdmin {
		facades.Log().Info("User sudah ada")
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "sudah terdaftar ",
		})
	}

	// Hash password
	hashedPassword, err := facades.Hash().Make(admin.Password)
	if err != nil {
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "gagal hashing password",
		})
	}

	// Simpan admin baru (AMAN)
	err = facades.Orm().Query().Table("admins").Create(map[string]any{
		"name":        admin.Name,
		"email":       admin.Email,
		"password":    hashedPassword,
		"code_rendem": admin.CodeRendem,
	})

	if err != nil {
		facades.Log().Error("Gagal menyimpan admin:", err)
		return ctx.Response().Json(http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Flash message & redirect ke login
	return ctx.Response().Redirect(http.StatusTemporaryRedirect, "/admin/signIn/")
}

func (r *AdminController) SettingsView(ctx http.Context) http.Response {

	var user models.Users

	err := facades.Auth(ctx).User(&user)
	if err != nil {
		log.Println("Error mendapatkan user : ", err)
		return ctx.Response().Json(http.StatusOK, map[string]any{
			"status":  "error",
			"message": "Gagal mendapatkaan users",
		})
	}
	return ctx.Response().View().Make("admin/settings.html", map[string]any{
		"title": fmt.Sprintf("settings | %S ", user.Name),
	})
}
