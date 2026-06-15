package controllers

import (
	"fmt"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/services"
)

type TelemetryController struct {
	telemetry *services.Telemetry
}

func NewTelemetryController() *TelemetryController {
	// Build the telemetry example once; its instruments are created in the
	// constructor and reused across requests.
	telemetry, err := services.NewTelemetry()
	if err != nil {
		facades.Log().Error(fmt.Sprintf("failed to build the telemetry example service: %+v", err))
	}

	return &TelemetryController{
		telemetry: telemetry,
	}
}

func (r *TelemetryController) Index(ctx http.Context) http.Response {
	// The whole endpoint is the manual telemetry example, so surface a failed
	// initialization instead of returning a misleading 200 with no telemetry.
	if r.telemetry == nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": "telemetry example is not initialized",
		})
	}

	facades.Log().Channel("otel").WithContext(ctx).Info("test telemetry log")

	resp, err := facades.Http().WithContext(ctx).Get("/grpc/user?token=1")
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	body, err := resp.Body()
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"error": err.Error(),
		})
	}

	// userID defaults to "1"; request ?user= (empty) to exercise the validation error path.
	userID := ctx.Request().Query("user", "1")
	if err := r.telemetry.Process(ctx, userID); err != nil {
		facades.Log().WithContext(ctx).Error("user processing failed: ", err)
	}
	r.telemetry.Consume(r.telemetry.Publish(ctx))

	return ctx.Response().Success().String(body)
}
