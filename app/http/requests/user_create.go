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

func (r *UserCreate) Rules() map[string]string {
	return map[string]string{
		"name": "required",
	}
}

func (r *UserCreate) Messages() map[string]string {
	return map[string]string{}
}

func (r *UserCreate) Attributes() map[string]string {
	return map[string]string{}
}

func (r *UserCreate) PrepareForValidation(data validation.Data) error {
	return nil
}
