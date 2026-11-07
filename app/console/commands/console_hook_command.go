package commands

import (
	"errors"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/console/hooks"
)

// ErrConsoleHookPublishFailed is returned by ConsoleHookCommand.Handle when the
// run is told to fail, so a test can assert the After hook receives and wraps it.
var ErrConsoleHookPublishFailed = errors.New("publish failed")

// ConsoleHookCommand is the command the WithCommandHooks tests run.
type ConsoleHookCommand struct{}

func (r *ConsoleHookCommand) Signature() string {
	return "test:console-hook"
}

func (r *ConsoleHookCommand) Description() string {
	return "Demonstrate WithCommandHooks around a release publish"
}

func (r *ConsoleHookCommand) Extend() command.Extend {
	return command.Extend{
		Category: "console",
		Flags: []command.Flag{
			&command.StringFlag{Name: "target", Value: "staging", Usage: "Publish target"},
			&command.BoolFlag{Name: "fail", Usage: "Return an error after publishing"},
		},
	}
}

func (r *ConsoleHookCommand) Handle(ctx console.Context) error {
	hooks.Record("handle")

	target := ctx.Option("target")
	ctx.Info("publishing to " + target)

	if ctx.OptionBool("fail") {
		return ErrConsoleHookPublishFailed
	}

	return nil
}
