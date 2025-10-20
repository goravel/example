package console

import (
	"goravel/app/console/commands"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/facades"
)

type Kernel struct {
}

func (kernel *Kernel) Schedule() []schedule.Event {
	return []schedule.Event{
		facades.Schedule().Command("app:test1").EverySecond(),
		facades.Schedule().Command("app:test2").EverySecond(),
		facades.Schedule().Command("app:test3").EverySecond(),
	}
}

func (kernel *Kernel) Commands() []console.Command {
	return []console.Command{
		&commands.Test1{},
		&commands.Test2{},
		&commands.Test3{},
		&commands.Test{},
	}
}
