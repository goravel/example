package providers

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
	"goravel/app/models"
	"goravel/app/observers"
)

type EventServiceProvider struct {
}

func (receiver *EventServiceProvider) Register(app foundation.Application) {
	facades.Event().Register(receiver.listen())
}

func (receiver *EventServiceProvider) Boot(app foundation.Application) {
	facades.Orm().Observe(models.User{}, &observers.UserObserver{})
}

func (receiver *EventServiceProvider) listen() map[event.Event][]event.Listener {
	return map[event.Event][]event.Listener{}
}
