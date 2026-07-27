package events

import (
	"time"

	"github.com/goravel/framework/broadcasting"
	contracts "github.com/goravel/framework/contracts/broadcasting"
)

type OrderShippedBroadcast struct {
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

func (e *OrderShippedBroadcast) BroadcastOn() []contracts.Channel {
	return []contracts.Channel{
		broadcasting.PrivateChannel("orders." + itoa(e.OrderID)),
	}
}

func (e *OrderShippedBroadcast) BroadcastAs() string {
	return "order.shipped"
}

func (e *OrderShippedBroadcast) BroadcastWith() map[string]any {
	return map[string]any{"order": e.OrderData}
}

func (e *OrderShippedBroadcast) BroadcastWhen() bool {
	return e.ShouldFire
}

func (e *OrderShippedBroadcast) BroadcastQueue() string {
	return e.QueueName
}

func (e *OrderShippedBroadcast) BroadcastConnections() []string {
	return e.Conns
}

func (e *OrderShippedBroadcast) BroadcastQueueConnection() string {
	return e.QueueConn
}

func (e *OrderShippedBroadcast) BroadcastDelay() time.Time {
	return e.DelayedAt
}

func (e *OrderShippedBroadcast) BroadcastTries() int {
	return e.Retries
}

func (e *OrderShippedBroadcast) BroadcastBackoff() time.Duration {
	return e.Backoff
}

func (e *OrderShippedBroadcast) BroadcastTimeout() time.Duration {
	return e.Timeout
}

type OrderShippedNowBroadcast struct {
	OrderID   uint
	OrderData map[string]any
}

func (e *OrderShippedNowBroadcast) BroadcastOn() []contracts.Channel {
	return []contracts.Channel{
		broadcasting.PublicChannel("orders"),
	}
}

func (e *OrderShippedNowBroadcast) BroadcastAs() string {
	return "order.shipped"
}

func (e *OrderShippedNowBroadcast) BroadcastWith() map[string]any {
	return map[string]any{"order": e.OrderData}
}

func (e *OrderShippedNowBroadcast) BroadcastWhen() bool {
	return true
}

func (e *OrderShippedNowBroadcast) BroadcastNow() bool {
	return true
}

type EmptyBroadcastEvent struct{}

func (e *EmptyBroadcastEvent) BroadcastOn() []contracts.Channel {
	return nil
}

func (e *EmptyBroadcastEvent) BroadcastAs() string {
	return "empty.event"
}

func (e *EmptyBroadcastEvent) BroadcastWith() map[string]any {
	return nil
}

func (e *EmptyBroadcastEvent) BroadcastWhen() bool {
	return false
}

type TeamPresenceBroadcast struct {
	TeamID   uint
	TeamData map[string]any
}

func (e *TeamPresenceBroadcast) BroadcastOn() []contracts.Channel {
	return []contracts.Channel{
		broadcasting.PresenceChannel("team." + itoa(e.TeamID)),
	}
}

func (e *TeamPresenceBroadcast) BroadcastAs() string {
	return "team.presence"
}

func (e *TeamPresenceBroadcast) BroadcastWith() map[string]any {
	return map[string]any{"team": e.TeamData}
}

func (e *TeamPresenceBroadcast) BroadcastWhen() bool {
	return true
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
