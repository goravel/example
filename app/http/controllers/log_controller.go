package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type logCtxKey string
type logSentinel struct{}

type LogController struct {
}

func NewLogController() *LogController {
	return &LogController{}
}

// WithContext demonstrates the v1.18 behavior of
// facades.Log().WithContext(ctx): framework-internal context keys and
// user-configured keys (logging.context.exclude) are filtered out of the
// emitted log entry, while unrelated keys are kept under their short names.
func (r *LogController) WithContext(ctx http.Context) http.Response {
	ctx.WithValue("GoravelAuthJwt", "should-be-filtered")
	ctx.WithValue("access_token", "should-be-filtered")
	ctx.WithValue("secret_key", "should-be-filtered")
	ctx.WithValue("request_id", "req-abc-123")
	ctx.WithValue(logCtxKey("trace_id"), "trace-xyz-987")
	ctx.WithValue(logSentinel{}, "user-42")

	facades.Log().WithContext(ctx).Info("log with context example")

	return ctx.Response().Success().Json(http.Json{
		"ok": true,
	})
}
