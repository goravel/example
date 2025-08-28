package requests

import (
	"errors"
	"goravel/app/models"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"
)

// UpdatePostRequest contains the validation rules for the blog post update request.
type UpdatePostRequest struct {
	Title         string    `form:"title"   json:"title"`
	PublishedDate time.Time `form:"publishDate"   json:"publishDate"`
	Author        string    `form:"author"   json:"author"`
	Body          string    `form:"body"   json:"body"`
}

// Authorize checks if the user is authorized to make this request.
func (r *UpdatePostRequest) Authorize(ctx http.Context) error {
	var post models.Post

	user := ctx.Value("user").(models.User)
	facades.Log().Debugf("(Request) user: %#v", user)

	id := ctx.Request().RouteInt("id")

	err := facades.Orm().Query().FindOrFail(&post, id)
	if err != nil {
		facades.Log().Errorf("failed to get post with id %d: %v", id, err)

		return errors.New("post not found")
	}

	if !facades.Gate().Allows("update-post", map[string]any{
		"post": post,
		"user": user,
	}) {
		return errors.New("can't update post")
	}

	return nil
}

// Filters applies filters to the request data.
func (r *UpdatePostRequest) Filters(_ http.Context) map[string]string {
	return map[string]string{
		"title":       "trim",
		"publishDate": "trim",
		"author":      "trim",
		"body":        "trim",
	}
}

// Rules returns the validation rules for the user registration request.
func (r *UpdatePostRequest) Rules(_ http.Context) map[string]string {
	return map[string]string{
		"title":       "required|string|min_len:3|max_len:255",
		"publishDate": "date",
		"author":      "required|string|min_len:3|max_len:255",
		"body":        "required|string|min_len:3",
	}
}

// Messages returns the custom validation error messages for the user registration request.
func (r *UpdatePostRequest) Messages(_ http.Context) map[string]string {
	return map[string]string{}
}

// Attributes returns the custom validation attributes to apply to the request.
func (r *UpdatePostRequest) Attributes(_ http.Context) map[string]string {
	return map[string]string{}
}

// PrepareForValidation is a hook that is called before validation.
func (r *UpdatePostRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}
