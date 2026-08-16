package notifications

import (
	"github.com/goravel/framework/contracts/notification"
)

type Welcome struct {
	Name string
}

func NewWelcome(name string) *Welcome {
	return &Welcome{Name: name}
}

func (r *Welcome) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase}
}

func (r *Welcome) ID() string {
	return "welcome-notification"
}

func (r *Welcome) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "Welcome " + r.Name}
}
