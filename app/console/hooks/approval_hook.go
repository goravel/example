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
// skips the command unless the flag is set. After is a pass-through, so the
// command's error reaches the caller untouched.
type ApprovalHook struct{}

var _ console.CommandHook = (*ApprovalHook)(nil)

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

	ctx.Comment("skipped: pass --approved to run")

	return false, nil
}

// After exists only because CommandHook requires it; the gate acts in Before.
func (r *ApprovalHook) After(_ console.Context, handleErr error) error {
	return handleErr
}

func (r *ApprovalHook) Flags() []command.Flag {
	return []command.Flag{
		&command.BoolFlag{
			Name:  "approved",
			Usage: "Allow the command to run",
		},
	}
}
