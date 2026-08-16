package notifications

import (
	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/notification/mail"
)

type OrderShipped struct {
	OrderID string
}

func NewOrderShipped(orderID string) *OrderShipped {
	return &OrderShipped{OrderID: orderID}
}

func (r *OrderShipped) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelMail}
}

func (r *OrderShipped) ToMail(notifiable notification.Notifiable) notification.MailMessage {
	return mail.NewMessage().
		Subject("Order " + r.OrderID + " has shipped").
		Html("<p>Order " + r.OrderID + " has shipped.</p>").
		Build()
}
