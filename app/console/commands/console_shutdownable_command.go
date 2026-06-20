package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type ConsoleShutdownableCommand struct{}

func (r *ConsoleShutdownableCommand) Signature() string {
	return "test:console-shutdownable"
}

func (r *ConsoleShutdownableCommand) Description() string {
	return "Demonstrate graceful shutdown via console.Shutdownable"
}

func (r *ConsoleShutdownableCommand) Extend() command.Extend {
	return command.Extend{Category: "console"}
}

func (r *ConsoleShutdownableCommand) Handle(_ console.Context) error {
	SetShutdownableHandleRan()
	return nil
}

func (r *ConsoleShutdownableCommand) Shutdown(_ console.Context) error {
	SetShutdownableShutdownRan()
	return nil
}
