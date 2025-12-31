package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
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
	
	return nil
}
