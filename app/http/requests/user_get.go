package requests

import (
	"github.com/goravel/framework/contracts/http"
)

type UserGet struct {
	Date []string `form:"date" json:"date"`
}

func (r *UserGet) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserGet) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		// "date":   "required|array|len:2",
		"date.*": "required|date",
	}
}
