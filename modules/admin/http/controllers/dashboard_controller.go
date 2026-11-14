package controllers

import (
	"encoding/json"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/modules/admin/dashboard"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// Show renders the page the admin mounts in, with every dashboard script in one
// Vite call so shared chunks are preloaded once.
func (r *DashboardController) Show(ctx http.Context) http.Response {
	entries, err := dashboard.Entries(facades.Modular())
	if err != nil {
		return r.fail(ctx, err)
	}

	tags, err := facades.Vite().Tags(entries...)
	if err != nil {
		return r.fail(ctx, err)
	}

	boot, err := json.Marshal(map[string]any{
		"appName":  facades.Config().GetString("app.name"),
		"basePath": "/" + strings.Trim(facades.Config().GetString("modules.admin.path", "admin"), "/"),
	})
	if err != nil {
		return r.fail(ctx, err)
	}

	return ctx.Response().View().Make("admin/dashboard.tmpl", map[string]any{
		"title": facades.Lang(ctx).Get("admin/messages.title"),
		"tags":  tags,
		"boot":  string(boot),
	})
}

func (r *DashboardController) fail(ctx http.Context, err error) http.Response {
	facades.Log().Error(err)

	return ctx.Response().String(http.StatusInternalServerError, err.Error())
}
