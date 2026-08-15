package facades

import (
	"github.com/goravel/framework/contracts/notification"
)

func Notification() notification.Manager {
	return App().MakeNotification()
}
