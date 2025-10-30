package requests

import (
	"goravel/app/models"

	"github.com/goravel/framework/contracts/http"
)

type UserCreate struct {
	Name   string           `form:"name" json:"name"`
	Avatar string           `form:"avatar" json:"avatar"`
	Alias  string           `form:"alias" json:"alias"`
	Mail   string           `form:"mail" json:"mail"`
	Tags   []models.UserTag `form:"tags" json:"tags"`
}

func (r *UserCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}
