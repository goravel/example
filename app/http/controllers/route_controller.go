package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

type RouteController struct {
	// Dependent services
}

func NewRouteController() *RouteController {
	return &RouteController{
		// Inject services
	}
}

func (r *RouteController) Throttle(ctx http.Context) http.Response {
	return ctx.Response().Success().String("success")
}
