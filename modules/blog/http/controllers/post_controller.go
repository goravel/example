package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/modules/blog/models"
)

type PostController struct{}

func NewPostController() *PostController {
	return &PostController{}
}

// Index returns the latest posts and the total, which the dashboard card shows.
func (r *PostController) Index(ctx http.Context) http.Response {
	var posts []models.Post
	if err := facades.Orm().Query().Order("id desc").Limit(facades.Config().GetInt("modules.blog.per_page")).Find(&posts); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": err.Error()})
	}

	total, err := facades.Orm().Query().Model(&models.Post{}).Count()
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": err.Error()})
	}

	return ctx.Response().Success().Json(http.Json{
		"data":  posts,
		"total": total,
	})
}

func (r *PostController) Store(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"title": "required|max_len:255",
		"body":  "required",
	})
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"message": err.Error()})
	}
	if validator.Fails() {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{"errors": validator.Errors().All()})
	}

	post := models.Post{
		Title: ctx.Request().Input("title"),
		Body:  ctx.Request().Input("body"),
	}
	if err := facades.Orm().Query().Create(&post); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": err.Error()})
	}

	return ctx.Response().Json(http.StatusCreated, http.Json{"data": post})
}
