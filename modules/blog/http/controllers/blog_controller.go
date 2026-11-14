package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/modules/blog/models"
)

type BlogController struct{}

func NewBlogController() *BlogController {
	return &BlogController{}
}

// Index is the public blog page: a module view with a module translation.
func (r *BlogController) Index(ctx http.Context) http.Response {
	var posts []models.Post
	if err := facades.Orm().Query().Order("id desc").Limit(facades.Config().GetInt("modules.blog.per_page")).Find(&posts); err != nil {
		return ctx.Response().String(http.StatusInternalServerError, err.Error())
	}

	return ctx.Response().View().Make("blog/index.tmpl", map[string]any{
		"title": facades.Lang(ctx).Get("blog/messages.title"),
		"posts": posts,
	})
}
