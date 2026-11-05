package hooks

import (
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

var _ console.CommandHook = (*AuditHook)(nil)

// AuditHook records what the command was asked to do and annotates a failed
// run with the hook's own context. It defines no flags, so it only reads what
// the command and the other hooks already add.
type AuditHook struct{}

func NewAuditHook() *AuditHook {
	return &AuditHook{}
}

func (r *AuditHook) Before(ctx console.Context) (bool, error) {
	Record("audit:before")
	SetTarget(ctx.Option("target"))

	return true, nil
}

func (r *AuditHook) After(_ console.Context, handleErr error) error {
	Record("audit:after")
	SetHandleErr(handleErr)

	if handleErr == nil {
		return nil
	}

	return fmt.Errorf("audit hook: %w", handleErr)
}

// Flags adds none: --target belongs to the command, not to the hook.
func (r *AuditHook) Flags() []command.Flag {
	return nil
}
