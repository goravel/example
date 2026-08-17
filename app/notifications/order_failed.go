package notifications

import (
	"time"

	"github.com/goravel/framework/contracts/notification"
)

// OrderFailed is a queued notification demonstrating NotificationWithTries
// and NotificationWithBackoff: if database-channel delivery fails
// transiently, the queued DispatchJob retries up to 3 times (1s before the
// second attempt, 2s before the third and any further attempts).
type OrderFailed struct {
	OrderID string
}

func NewOrderFailed(orderID string) *OrderFailed {
	return &OrderFailed{OrderID: orderID}
}

func (r *OrderFailed) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase}
}

func (r *OrderFailed) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"order_id": r.OrderID, "status": "failed"}
}

// OnQueue/OnConnection implement contracts/notification.ShouldQueue, marking
// this notification for queued dispatch (default queue/connection).
func (r *OrderFailed) OnQueue() string      { return "" }
func (r *OrderFailed) OnConnection() string { return "" }

// Tries implements NotificationWithTries: allow up to 3 delivery attempts.
func (r *OrderFailed) Tries(channel string) int { return 3 }

// Backoff implements NotificationWithBackoff: 1s before attempt 2, 2s
// before attempt 3+.
func (r *OrderFailed) Backoff(channel string) []time.Duration {
	return []time.Duration{time.Second, 2 * time.Second}
}
