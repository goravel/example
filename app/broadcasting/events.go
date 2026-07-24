package broadcasting

import (
	"time"

	"github.com/goravel/framework/broadcasting"
	contracts "github.com/goravel/framework/contracts/broadcasting"
)

type OrderShipped struct {
	OrderID    uint
	OrderData  map[string]any
	ShouldFire bool
	QueueName  string
	Conns      []string
	QueueConn  string
	DelayedAt  time.Time
	Retries    int
	Backoff    time.Duration
	Timeout    time.Duration
}

func (e *OrderShipped) BroadcastOn() []contracts.Channel {
	return []contracts.Channel{
		broadcasting.PrivateChannel("orders." + itoa(e.OrderID)),
	}
}

func (e *OrderShipped) BroadcastAs() string {
	return "order.shipped"
}

func (e *OrderShipped) BroadcastWith() map[string]any {
	return map[string]any{"order": e.OrderData}
}

func (e *OrderShipped) BroadcastWhen() bool {
	return e.ShouldFire
}

func (e *OrderShipped) BroadcastQueue() string {
	return e.QueueName
}

func (e *OrderShipped) BroadcastConnections() []string {
	return e.Conns
}

func (e *OrderShipped) BroadcastQueueConnection() string {
	return e.QueueConn
}

func (e *OrderShipped) BroadcastDelay() time.Time {
	return e.DelayedAt
}

func (e *OrderShipped) BroadcastTries() int {
	return e.Retries
}

func (e *OrderShipped) BroadcastBackoff() time.Duration {
	return e.Backoff
}

func (e *OrderShipped) BroadcastTimeout() time.Duration {
	return e.Timeout
}

type OrderShippedNow struct {
	OrderID   uint
	OrderData map[string]any
}

func (e *OrderShippedNow) BroadcastOn() []contracts.Channel {
	return []contracts.Channel{
		broadcasting.PublicChannel("orders"),
	}
}

func (e *OrderShippedNow) BroadcastAs() string {
	return "order.shipped"
}

func (e *OrderShippedNow) BroadcastWith() map[string]any {
	return map[string]any{"order": e.OrderData}
}

func (e *OrderShippedNow) BroadcastWhen() bool {
	return true
}

func (e *OrderShippedNow) BroadcastNow() bool {
	return true
}

type EmptyEvent struct{}

func (e *EmptyEvent) BroadcastOn() []contracts.Channel {
	return nil
}

func (e *EmptyEvent) BroadcastAs() string {
	return "empty.event"
}

func (e *EmptyEvent) BroadcastWith() map[string]any {
	return nil
}

func (e *EmptyEvent) BroadcastWhen() bool {
	return false
}

func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
