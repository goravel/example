package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

type TestController struct {
}

func NewTestController() *TestController {
	return &TestController{
		// Inject any required services
	}
}

// Index handles a basic GET request
func (r *TestController) Index(ctx http.Context) http.Response {
	return ctx.Response().Json(http.StatusOK, map[string]string{"message": "Welcome to the Index route"})
}

// Show handles a GET request for a specific item by ID
func (r *TestController) Show(ctx http.Context) http.Response {
	id := ctx.Request().Route("id")
	if id == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"error": "ID is required"})
	}
	return ctx.Response().Json(http.StatusOK, map[string]string{"id": id, "message": "Item retrieved successfully"})
}

// Create handles a POST request to create a new item
func (r *TestController) Create(ctx http.Context) http.Response {
	var data map[string]any
	if err := ctx.Request().Bind(&data); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}
	return ctx.Response().Json(http.StatusCreated, map[string]interface{}{"message": "Item created", "data": data})
}

// Update handles a PUT request to update an existing item
func (r *TestController) Update(ctx http.Context) http.Response {
	id := ctx.Request().Route("id")
	var data map[string]interface{}
	if err := ctx.Request().Bind(&data); err != nil || id == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	return ctx.Response().Json(http.StatusOK, map[string]interface{}{"message": "Item updated", "id": id, "data": data})
}

// Delete handles a DELETE request to delete an item by ID
func (r *TestController) Delete(ctx http.Context) http.Response {
	id := ctx.Request().Route("id")
	if id == "" {
		return ctx.Response().Json(http.StatusBadRequest, map[string]string{"error": "ID is required"})
	}
	return ctx.Response().Json(http.StatusOK, map[string]string{"message": "Item deleted", "id": id})
}

// CustomHeader handles a GET request and returns a response with custom headers
func (r *TestController) CustomHeader(ctx http.Context) http.Response {
	return ctx.Response().Header("X-Custom-Header", "CustomHeaderValue").
		Json(http.StatusOK, map[string]string{"message": "Response with custom header"})
}

// EmptyResponse returns a no-content response (204 status)
func (r *TestController) EmptyResponse(ctx http.Context) http.Response {
	return ctx.Response().NoContent()
}
