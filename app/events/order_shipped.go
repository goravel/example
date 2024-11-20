package events

import "github.com/goravel/framework/contracts/event"

type OrderShipped struct {
}

func NewOrderShipped() *OrderShipped {
	return &OrderShipped{}
}

func (receiver *OrderShipped) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}
