package hooks

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// tracePrefix is the command family the trace hook records. The hook is
// registered on the "*" key, so it runs for every Artisan command in the app
// and must stay a no-op outside this family.
const tracePrefix = "test:console-hook"

var _ console.CommandHook = (*TraceHook)(nil)

// TraceHook records the outermost Before and After markers around
// test:console-hook. Flags stays empty on purpose: a "*" hook that added a flag
// would add it to every command in the app.
type TraceHook struct{}

func NewTraceHook() *TraceHook {
	return &TraceHook{}
}

func (r *TraceHook) Before(ctx console.Context) (bool, error) {
	if !r.traced(ctx) {
		return true, nil
	}

	Record("trace:before")

	return true, nil
}

func (r *TraceHook) After(ctx console.Context, handleErr error) error {
	if !r.traced(ctx) {
		return handleErr
	}

	Record("trace:after")

	return handleErr
}

// Flags adds none: a "*" hook that defined a flag would add it to every
// command in the app.
func (r *TraceHook) Flags() []command.Flag {
	return nil
}

func (r *TraceHook) traced(ctx console.Context) bool {
	return strings.HasPrefix(ctx.Instance().Name, tracePrefix)
}
