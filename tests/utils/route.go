package utils

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/goravel/framework/facades"
)

func Http() *resty.Client {
	return resty.New().
		SetBaseURL(fmt.Sprintf("http://%s:%s",
			facades.Config().GetString("APP_HOST"),
			facades.Config().GetString("APP_PORT"))).
		SetHeader("Content-Type", "application/json")
}
