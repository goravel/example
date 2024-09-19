package controllers

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

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

func (r *UserController) Store(ctx http.Context) http.Response {
	user := models.User{
		Name:   ctx.Request().Input("name"),
		Avatar: ctx.Request().Input("avatar"),
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

// Save
//
//	@Description: 保存
//	@receiver r
//	@param ctx
//	@return http.Response
func (r *UserController) Save(ctx http.Context) http.Response {

	user := models.User{
		Name:   ctx.Request().Input("name"),
		Avatar: ctx.Request().Input("avatar"),
	}
	err := facades.Orm().Transaction(func(tx orm.Transaction) error {
		err := tx.FirstOrNew(&models.User{}, models.User{ID: user.ID})
		if err != nil {
			return err
		}
		err = tx.Save(&user)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}
	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}
