package controllers

import (
	"context"

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
	reqCtx := context.WithValue(ctx.Context(), "GoravelAuthJwt", "should-be-filtered")
	reqCtx = context.WithValue(reqCtx, "access_token", "should-be-filtered")
	reqCtx = context.WithValue(reqCtx, "secret_key", "should-be-filtered")
	reqCtx = context.WithValue(reqCtx, "request_id", "req-abc-123")
	reqCtx = context.WithValue(reqCtx, logCtxKey("trace_id"), "trace-xyz-987")
	reqCtx = context.WithValue(reqCtx, logSentinel{}, "user-42")

	facades.Log().WithContext(reqCtx).Info("log with context example")

	return ctx.Response().Success().Json(http.Json{
		"ok": true,
	})
}
