package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/spf13/cast"
)

var TestResultOfSendShipmentNotification []string

type SendShipmentNotification struct {
}

func NewSendShipmentNotification() *SendShipmentNotification {
	return &SendShipmentNotification{}
}

func (receiver *SendShipmentNotification) Signature() string {
	return "send_shipment_notification"
}

func (receiver *SendShipmentNotification) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *SendShipmentNotification) Handle(args ...any) error {
	if len(args) > 0 {
		TestResultOfSendShipmentNotification = append(TestResultOfSendShipmentNotification, cast.ToString(args[0]))
	}

	return nil
}
