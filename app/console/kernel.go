package console

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/facades"
)

type Kernel struct {
}

func (kernel *Kernel) Schedule() []schedule.Event {
	return []schedule.Event{

		facades.Schedule().Call(func() {
			fmt.Println("开始执行计划任务,10:19")
		}).DailyAt("10:19"),
		facades.Schedule().Call(func() {
			fmt.Println(time.Now())
			fmt.Println("开始执行计划任务，每分钟执行一次")
		}).EveryMinute(),
	}
}

func (kernel *Kernel) Commands() []console.Command {
	return []console.Command{}
}
