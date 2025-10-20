package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type Test2 struct {
}

// Signature The name and signature of the console command.
func (r *Test2) Signature() string {
	return "app:test2"
}

// Description The console command description.
func (r *Test2) Description() string {
	return "Command description"
}

// Extend The console command extend.
func (r *Test2) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle Execute the console command.
func (r *Test2) Handle(ctx console.Context) error {
	facades.Log().Info("app:test2 start")
	facades.Log().Info("app:test2 end")
	return nil
}
