package routes

import (
	"goravel/app/facades"
	"goravel/app/models"

	"github.com/spf13/cast"
)

func Channels() {
	facades.Broadcast().Channel("orders.{orderId}", func(userID any, channelName string, params map[string]string) (bool, any) {
		return userID != nil && params["orderId"] != "", nil
	})

	facades.Broadcast().Channel("users.{userId}", func(userID any, channelName string, params map[string]string) (bool, any) {
		var user models.User
		if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil {
			return false, nil
		}

		return params["userId"] == cast.ToString(user.ID), &user
	})

	facades.Broadcast().Channel("public", func(userID any, channelName string, params map[string]string) (bool, any) {
		return true, nil
	})

	facades.Broadcast().Channel("team.{teamId}", func(userID any, channelName string, params map[string]string) (bool, any) {
		return userID != nil && params["teamId"] != "", nil
	})
}
