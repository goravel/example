package commands

import (
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type Test struct {
}

// Signature The name and signature of the console command.
func (r *Test) Signature() string {
	return "app:test"
}

// Description The console command description.
func (r *Test) Description() string {
	return "Command description"
}

// Extend The console command extend.
func (r *Test) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle Execute the console command.
func (r *Test) Handle(ctx console.Context) error {
	go facades.Artisan().Call("app:test1")
	go facades.Artisan().Call("app:test2")
	go facades.Artisan().Call("app:test3")
	time.Sleep(1 * time.Second)
	return nil
}
