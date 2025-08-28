package admin

import (
	"bytes"
	"html/template"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"goravel/app/http/requests/admin/post"
	"goravel/app/models"
)

// PagerCount is the number of posts to display per page.
const PagerCount = 10

// PostController is the controller for the admin blog post routes.
type PostController struct {
	// Dependent services
}

// NewPostController creates a new instance of the PostController.
func NewPostController() *PostController {
	return &PostController{
		// Inject services
	}
}

// Index handles the blog post index route.
func (r *PostController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("admin/posts/index.tmpl")
}

// GetPostsWidget returns the blog post widget.
func (r *PostController) GetPostsWidget(ctx http.Context) http.Response {
	page := ctx.Request().QueryInt("page", 1)

	facades.Log().Debugf("page: %d", page)

	var (
		posts []models.Post
		total int64
	)

	err := facades.Orm().Query().With("User").Paginate(page, PagerCount, &posts, &total)
	if err != nil {
		facades.Log().Errorf("failed to get posts: %v", err)

		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	facades.Log().Debugf("post total: %d, posts: %#v", total, posts)

	type Posts struct {
		Posts      []models.Post
		Pagination map[string]int
	}

	ps := Posts{
		Posts: posts,
		Pagination: map[string]int{
			"prev":    page - 1,
			"current": page,
			"next":    page + 1,
		},
	}

	return ctx.Response().View().Make("admin/posts/widgets/list.tmpl", map[string]any{
		"Posts": ps,
		"Total": total / int64(PagerCount),
	})
}

// Show handles the blog post show route.
func (r *PostController) Show(ctx http.Context) http.Response {
	postID := ctx.Request().RouteInt("id")

	facades.Log().Debugf("postID: %d", postID)

	var post models.Post

	err := facades.Orm().Query().With("User").Where("id", postID).FirstOrFail(&post)
	if err != nil {
		facades.Log().Errorf("failed to get post: %v", err)

		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	facades.Log().Debugf("post: %#v", post)

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer

	err = md.Convert([]byte(post.Body), &buf)
	if err != nil {
		facades.Log().Errorf("failed to convert markdown to html: %v", err)

		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	type Post struct {
		ID            uint
		Title, Author string
		Date          time.Time
		Body          template.HTML
	}

	return ctx.Response().View().Make("admin/posts/post.tmpl", map[string]any{
		"Admin": true,
		"Post": Post{
			ID:     post.ID,
			Title:  post.Title,
			Date:   post.PublishedAt,
			Author: post.User.Username,
			Body:   template.HTML(buf.String()), //nolint: gosec
		},
	})
}

// Store handles the blog post store route.
func (r *PostController) Store(_ http.Context) http.Response {
	return nil
}

// Update handles the blog post update route.
func (r *PostController) Update(ctx http.Context) http.Response {
	postID := ctx.Request().RouteInt("id")

	facades.Log().Infof("updating blog post with id: %d", postID)

	if ctx.Request().Method() == http.MethodGet {
		var post models.Post

		err := facades.Orm().Query().With("User").Where("id", postID).FirstOrFail(&post)
		if err != nil {
			facades.Log().Errorf("failed to get post: %v", err)

			return ctx.Response().Json(http.StatusInternalServerError, http.Json{
				"error": err.Error(),
			})
		}

		facades.Log().Debugf("post: %#v", post)

		type Post struct {
			ID                  uint
			Title, Author, Body string
			PublishedAt         time.Time
			Date                time.Time
		}

		return ctx.Response().View().Make("admin/posts/edit.tmpl", map[string]any{
			"Admin": true,
			"Post": Post{
				ID:          post.ID,
				Title:       post.Title,
				Date:        post.CreatedAt.StdTime(),
				PublishedAt: post.PublishedAt,
				Author:      post.User.Username,
				Body:        post.Body,
			},
		})
	}

	var request post.UpdatePostRequest

	user := ctx.Value("user").(models.User)
	facades.Log().Debugf("(Controller) user: %#v", user)

	errs, err := ctx.Request().ValidateRequest(&request)
	if err != nil {
		facades.Log().Errorf("failed to validate request: %v", err)

		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	} else if errs != nil {
		facades.Log().Warningf("post update validation failed: %#v", errs.All())

		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{
			"error": errs.All(),
		})
	}

	facades.Log().Debugf("request: %#v", request)

	return nil
}

// Destroy handles the blog post destroy route.
func (r *PostController) Destroy(ctx http.Context) http.Response {
	postID := ctx.Request().RouteInt("id")

	facades.Log().Infof("deleting blog post with id: %d", postID)

	_, err := facades.Orm().Query().Model(&models.Post{}).Where("id", postID).Delete()
	if err != nil {
		facades.Log().Errorf("failed to delete blog post: %v", err)

		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	return r.Index(ctx)
}
