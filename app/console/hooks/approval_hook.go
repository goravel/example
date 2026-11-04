package hooks

import (
	"errors"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// ErrApprovalRequired is returned by ApprovalHook.Before when a production run
// arrives without --approved. Before returning an error skips Handle and every
// After, and the error becomes the command's result.
var ErrApprovalRequired = errors.New("publishing to production requires --approved")

// ApprovalHook is a confirmation-style gate: it adds an --approved flag and
// skips the command unless the flag is set. It embeds BaseCommandHook, so After
// passes the command's error through untouched.
type ApprovalHook struct {
	console.BaseCommandHook
}

func NewApprovalHook() *ApprovalHook {
	return &ApprovalHook{}
}

func (r *ApprovalHook) Before(ctx console.Context) (bool, error) {
	Record("approval:before")

	approved := ctx.OptionBool("approved")
	SetApproved(approved)

	if approved {
		return true, nil
	}

	if ctx.Option("target") == "production" {
		return false, ErrApprovalRequired
	}

	return false, nil
}

func (r *ApprovalHook) Flags() []command.Flag {
	return []command.Flag{
		&command.BoolFlag{
			Name:  "approved",
			Usage: "Allow the command to run",
		},
	}
}
