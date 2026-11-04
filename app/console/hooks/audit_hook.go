package hooks

import (
	"fmt"

	"github.com/goravel/framework/contracts/console"
)

// AuditHook records what the command was asked to do and annotates a failed
// run with the hook's own context.
type AuditHook struct {
	console.BaseCommandHook
}

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
