package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"
)

type ValidationCreate struct {
	Context string           `form:"context" json:"context"`
	Name    string           `form:"name" json:"name"`
	Tags    []string         `form:"tags" json:"tags"`
	Scores  []int            `form:"scores" json:"scores"`
	Items   []ValidationItem `form:"items" json:"items"`
	Meta    map[string]any   `form:"meta" json:"meta"`
	Date    carbon.Carbon    `form:"date" json:"date"`
	Code    int              `form:"code" json:"code"`
	Age     int              `form:"age" json:"age"`
}

type ValidationItem struct {
	Name string `form:"name" json:"name"`
}

func (r *ValidationCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *ValidationCreate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":         "required",
		"context":      "required",
		"tags.*":       "required|string",
		"scores.*":     "required|int",
		"items.*.name": "sometimes|required|string",
		"meta":         "sometimes|map",
		"meta.name":    "sometimes|required|string",
		"date":         "required|date",
		"code":         `required|regex:^\d{4,6}$`,
	}
}

func (r *ValidationCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *ValidationCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *ValidationCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	if scores, exist := data.Get("scores"); exist {
		if err := data.Set("scores", cast.ToIntSlice(scores)); err != nil {
			return err
		}
	}
	if c, exist := data.Get("context"); exist {
		// Test getting value from context: ValidationController.Request
		if err := data.Set("context", c.(string)+"_"+ctx.Value("ctx").(string)); err != nil {
			return err
		}
	}

	return nil
}

func (r *ValidationCreate) Filters(ctx http.Context) map[string]any {
	return map[string]any{
		"name": "trim",
	}
}
