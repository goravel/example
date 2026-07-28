package events

import (
	"time"

	"github.com/goravel/framework/broadcasting"
	contracts "github.com/goravel/framework/contracts/broadcasting"
)

type OrderShippedBroadcast struct {
	OrderData          map[string]any
	ShouldFire         bool
	QueueName          string
	Conns              []string
	QueueConn          string
	DelayedAt          time.Time
	Retries            int
	Backoff            time.Duration
	Timeout            time.Duration
	ShouldBroadcastNow bool
	ChannelType        string
	ChannelName        string
}

func (e *OrderShippedBroadcast) BroadcastOn() []contracts.Channel {
	if e.ChannelType == "public" {
		return []contracts.Channel{
			broadcasting.PublicChannel(e.ChannelName),
		}
	}
	return []contracts.Channel{
		broadcasting.PrivateChannel(e.ChannelName),
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

func (e *OrderShippedBroadcast) BroadcastNow() bool {
	return e.ShouldBroadcastNow
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
	TeamData    map[string]any
	ChannelName string
}

func (e *TeamPresenceBroadcast) BroadcastOn() []contracts.Channel {
	return []contracts.Channel{
		broadcasting.PresenceChannel(e.ChannelName),
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
