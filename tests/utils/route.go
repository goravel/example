package utils

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/goravel/framework/facades"
)

var request *resty.Request

func Http() *resty.Request {
	if request == nil {
		request = resty.New().
			SetBaseURL(fmt.Sprintf("http://%s:%s",
				facades.Config().GetString("APP_HOST"),
				facades.Config().GetString("APP_PORT"))).
			SetHeader("Content-Type", "application/json").R()
	}

	return request
}
