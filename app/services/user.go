package services

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type User interface {
	Create(name string) (models.User, error)
}

type UserImpl struct {
}

func NewUserImpl() *UserImpl {
	return &UserImpl{}
}

func (r *UserImpl) Create(name string) (models.User, error) {
	user := models.User{
		Name:   name,
		Avatar: "avatar",
	}
	if err := facades.Orm().Query().Create(&user); err != nil {
		return user, err
	}

	return user, nil
}
