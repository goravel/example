package routes

import (
	"goravel/app/facades"

	"github.com/spf13/cast"
)

func Channels() {
	facades.Broadcast().Channel("orders.{orderId}", func(user any, channelName string, params map[string]string) bool {
		return user != nil && params["orderId"] != ""
	})

	facades.Broadcast().Channel("users.{userId}", func(user any, channelName string, params map[string]string) bool {
		userID := cast.ToString(user)
		return params["userId"] == userID
	})

	facades.Broadcast().Channel("public-updates", func(user any, channelName string, params map[string]string) bool {
		return true
	})
}
