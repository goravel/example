package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type Test3 struct {
}

// Signature The name and signature of the console command.
func (r *Test3) Signature() string {
	return "app:test3"
}

// Description The console command description.
func (r *Test3) Description() string {
	return "Command description"
}

// Extend The console command extend.
func (r *Test3) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle Execute the console command.
func (r *Test3) Handle(ctx console.Context) error {
	facades.Log().Info("app:test3 start")
	facades.Log().Info("app:test3 end")
	return nil
}
