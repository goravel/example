package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/hooks"
)

// CommandHooks returns the hooks that run around Artisan commands, keyed by
// command signature or '*' glob and matched like CommandsFilter. Keys apply in
// sorted order: "*" wraps "test:console-hook", which wraps "test:console-hook*".
// Before runs first to last, After last to first.
func CommandHooks() map[string][]console.CommandHook {
	return map[string][]console.CommandHook{
		// Runs for every command, so it stays a no-op outside the
		// test:console-hook family and adds no flags.
		"*": {hooks.NewTraceHook()},
		// Exact signature: the audit hook sits between the two glob keys.
		"test:console-hook": {hooks.NewAuditHook()},
		// Prefix glob: matches the same command and contributes --approved.
		"test:console-hook*": {hooks.NewApprovalHook()},
	}
}
