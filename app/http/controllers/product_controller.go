package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

type ProductController struct {
	// Dependent services
}

func NewProductController() *ProductController {
	return &ProductController{
		// Inject services
	}
}

func (r *ProductController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("product/index.html", map[string]any{
		"title": "penjualan",
	})
}
