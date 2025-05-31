package feature

import (
	"testing"
	"time"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/assert"

	"goravel/app/events"
	"goravel/app/listeners"
)

func TestEvent(t *testing.T) {
	assert.NoError(t, facades.Event().Job(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "I'm OrderShipped"},
	}).Dispatch())

	assert.NoError(t, facades.Event().Job(&events.OrderCanceled{}, []event.Arg{
		{Type: "string", Value: "I'm OrderCanceled"},
	}).Dispatch())

	time.Sleep(1 * time.Second)

	assert.ElementsMatch(t, []string{
		"I'm OrderShipped",
		"I'm OrderCanceled",
	}, listeners.TestResultOfSendShipmentNotification)
}
