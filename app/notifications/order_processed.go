package notifications

import (
	"github.com/goravel/framework/contracts/notification"
)

type OrderProcessed struct {
	OrderID string
}

func NewOrderProcessed(orderID string) *OrderProcessed {
	return &OrderProcessed{OrderID: orderID}
}

func (r *OrderProcessed) Via(notifiable notification.Notifiable) []string {
	return []string{"database"}
}

func (r *OrderProcessed) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"order_id": r.OrderID}
}

// OnQueue/OnConnection implement contracts/notification.ShouldQueue, marking
// this notification for queued dispatch (default queue/connection).
func (r *OrderProcessed) OnQueue() string      { return "" }
func (r *OrderProcessed) OnConnection() string { return "" }
