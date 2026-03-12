package controllers

import (
	"sync"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
	"goravel/app/http/requests"
	"goravel/app/models"
)

type UserController struct {
	//Dependent services
}

func NewUserController() *UserController {
	return &UserController{
		//Inject services
	}
}

// Index user index
// @Summary user index
// @Description user index
// @Tags User
// @Accept json
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /user [get]
func (r *UserController) Index(ctx http.Context) http.Response {
	var users []models.User
	if err := facades.Orm().Query().Get(&users); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"users": users,
	})
}

// Show user show
// @Summary user show
// @Description user show
// @Tags User
// @Accept json
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /user/{id} [get]
func (r *UserController) Show(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

// Store user store
// @Summary user store
// @Description user store
// @Tags User
// @Accept application/x-www-form-urlencoded,json,multipart/form-data
// @Param UserCreate formData requests.UserCreate false "requests.UserCreate data"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /user [post]
func (r *UserController) Store(ctx http.Context) http.Response {
	var userCreate requests.UserCreate
	errors, err := ctx.Request().ValidateRequest(&userCreate)
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}
	if errors != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": errors.All(),
		})
	}

	user := models.User{
		Name:   userCreate.Name,
		Avatar: userCreate.Avatar,
		Alias:  userCreate.Alias,
		Mail:   userCreate.Mail,
		Tags:   userCreate.Tags,
	}
	if err := facades.Orm().Query().Create(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

// Update user update
// @Summary user update
// @Description user update
// @Tags User
// @Accept json
// @Param name body string false "name"
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /user/{id} [put]
// @Router /user/{id} [patch]
func (r *UserController) Update(ctx http.Context) http.Response {
	if _, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Update(models.User{
		Name: ctx.Request().Input("name"),
	}); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

// Destroy user destroy
// @Summary user destroy
// @Description user destroy
// @Tags User
// @Accept json
// @Param id path string true "id"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /user/{id} [delete]
func (r *UserController) Destroy(ctx http.Context) http.Response {
	result, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Delete(&models.User{})
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"rows_affected": result.RowsAffected,
	})
}

var (
	UserControllerSingleton *UserController
	userControllerOnce      sync.Once
)

func (r *UserController) Singleton() *UserController {

	userControllerOnce.Do(func() {
		UserControllerSingleton = NewUserController()
	})

	return UserControllerSingleton
}

// Routes User routes.
// Example Usage:
// @api|web.go: controllers.UserControllerSingleton.Routes(nil)
func (r *UserController) Routes(baseRouter route.Router) {
	r.Singleton()
	var UserRouter = baseRouter
	if UserRouter == nil {
		UserRouter = facades.Route()
	}
	UserRouter.
		Get(
			"/user",

			UserControllerSingleton.
				Index)
	UserRouter.
		Get(
			"/user/{id}",

			UserControllerSingleton.
				Show)
	UserRouter.
		Post(
			"/user",

			UserControllerSingleton.
				Store)
	UserRouter.
		Put(
			"/user/{id}",

			UserControllerSingleton.
				Update)
	UserRouter.
		Patch(
			"/user/{id}",

			UserControllerSingleton.
				Update)
	UserRouter.
		Delete(
			"/user/{id}",

			UserControllerSingleton.
				Destroy)

}
