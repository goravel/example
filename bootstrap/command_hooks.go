package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/hooks"
)

// CommandHooks returns the hooks that run around Artisan commands, keyed by
// command signature or '*' glob and matched like CommandsFilter. The broadest
// key is outermost: "*" wraps "test:console-hook*", which wraps the exact
// "test:console-hook". Before runs outermost first, After innermost first.
func CommandHooks() map[string][]console.CommandHook {
	return map[string][]console.CommandHook{
		// Runs for every command, so it stays a no-op outside the
		// test:console-hook family and adds no flags.
		"*": {hooks.NewTraceHook()},
		// Exact signature: the audit hook sits innermost, after the glob.
		"test:console-hook": {hooks.NewAuditHook()},
		// Prefix glob: matches the same command and contributes --approved.
		"test:console-hook*": {hooks.NewApprovalHook()},
	}
}
