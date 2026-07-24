package routes

import (
	"goravel/app/facades"
)

func Channels() {
	facades.Broadcast().Channel("orders.{orderId}", func(user any, channelName string, params map[string]string) bool {
		userID := user.(map[string]any)["id"]
		return userID != nil && params["orderId"] != ""
	})

	facades.Broadcast().Channel("users.{userId}", func(user any, channelName string, params map[string]string) bool {
		userID := user.(map[string]any)["id"]
		return params["userId"] == userID
	})

	facades.Broadcast().Channel("public-updates", func(user any, channelName string, params map[string]string) bool {
		return true
	})
}
