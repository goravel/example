package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type Test1 struct {
}

// Signature The name and signature of the console command.
func (r *Test1) Signature() string {
	return "app:test1"
}

// Description The console command description.
func (r *Test1) Description() string {
	return "Command description"
}

// Extend The console command extend.
func (r *Test1) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle Execute the console command.
func (r *Test1) Handle(ctx console.Context) error {
	facades.Log().Info("app:test1 start")
	facades.Log().Info("app:test1 end")
	return nil
}
