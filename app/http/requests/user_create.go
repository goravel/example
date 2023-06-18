package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserCreate struct {
	Name string `form:"name" json:"name"`
}

func (r *UserCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *UserCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *UserCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *UserCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
