package events

import "github.com/goravel/framework/contracts/event"

type OrderCanceled struct {
}

func NewOrderCanceled() *OrderCanceled {
	return &OrderCanceled{}
}

func (receiver *OrderCanceled) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}
