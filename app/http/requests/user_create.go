package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"
)

type UserCreate struct {
	Name   string        `form:"name" json:"name"`
	Tags   []string      `form:"tags" json:"tags"`
	Scores []int         `form:"scores" json:"scores"`
	Date   carbon.Carbon `form:"date" json:"date"`
}

func (r *UserCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":     "required",
		"tags.*":   "required|string",
		"scores.*": "required|int",
		"date":     "required|date",
	}
}

func (r *UserCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *UserCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *UserCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	if scores, exist := data.Get("scores"); exist {
		return data.Set("scores", cast.ToIntSlice(scores))
	}

	return nil
}
