package services

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/goravel/framework/telemetry"

	"goravel/app/facades"
)

// scopeName identifies who produced the spans and metrics (the instrumentation
// scope). Convention is the instrumenting package or application name.
const scopeName = "goravel"

// errUserIDRequired is a simulated validation failure used to show error
// recording on a span.
var errUserIDRequired = errors.New("user id is required")

// Telemetry is a manual instrumentation example. It owns a tracer and a set of
// metric instruments created once and reused across requests.
type Telemetry struct {
	tracer    trace.Tracer
	processed metric.Int64Counter
	duration  metric.Float64Histogram
	inFlight  metric.Int64UpDownCounter
}

// NewTelemetry creates the tracer and metric instruments once. Instruments are
// safe for concurrent use and meant to be created once and reused, so build this
// once (see TelemetryController) and share it rather than rebuilding per request.
func NewTelemetry() (*Telemetry, error) {
	meter := facades.Telemetry().Meter(scopeName)

	processed, err := meter.Int64Counter("users.processed",
		metric.WithDescription("Number of processed users by result"))
	if err != nil {
		return nil, err
	}

	duration, err := meter.Float64Histogram("user.process.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of user processing"))
	if err != nil {
		return nil, err
	}

	inFlight, err := meter.Int64UpDownCounter("users.in_flight",
		metric.WithDescription("Users currently being processed"))
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		tracer:    facades.Telemetry().Tracer(scopeName),
		processed: processed,
		duration:  duration,
		inFlight:  inFlight,
	}, nil
}

// Process is a manually traced unit of work. It enriches the active (HTTP
// server) span, starts its own span with an attribute and an event, runs a
// nested child span, records the outcome on the counter and histogram, tracks
// in-flight work with the up-down counter, and logs with the trace context.
func (r *Telemetry) Process(ctx context.Context, userID string) error {
	start := time.Now()

	// Enrich the span already active on the incoming context (started by the
	// HTTP server middleware) before opening our own.
	trace.SpanFromContext(ctx).SetAttributes(telemetry.String("user.id", userID))

	ctx, span := r.tracer.Start(ctx, "users.process", telemetry.WithSpanKind(telemetry.SpanKindInternal))
	defer span.End()
	span.SetAttributes(telemetry.String("user.id", userID))

	r.inFlight.Add(ctx, 1)
	defer r.inFlight.Add(ctx, -1)

	if err := r.validate(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(telemetry.CodeError, err.Error())
		r.processed.Add(ctx, 1, metric.WithAttributes(telemetry.String("result", "error")))
		r.duration.Record(ctx, time.Since(start).Seconds())

		return err
	}

	span.AddEvent("user.validated")
	r.processed.Add(ctx, 1, metric.WithAttributes(telemetry.String("result", "ok")))
	r.duration.Record(ctx, time.Since(start).Seconds())
	facades.Log().WithContext(ctx).Info("user processed")

	return nil
}

// Publish injects the active trace context into message headers so a consumer
// on another transport can continue the same trace.
func (r *Telemetry) Publish(ctx context.Context) map[string]string {
	headers := map[string]string{}
	facades.Telemetry().Propagator().Inject(ctx, telemetry.PropagationMapCarrier(headers))

	return headers
}

// Consume extracts the propagated context and starts a consumer span that
// continues the producer's trace across the boundary.
func (r *Telemetry) Consume(headers map[string]string) {
	ctx := facades.Telemetry().Propagator().Extract(context.Background(), telemetry.PropagationMapCarrier(headers))

	ctx, span := r.tracer.Start(ctx, "users.consume", telemetry.WithSpanKind(telemetry.SpanKindConsumer))
	defer span.End()

	facades.Log().WithContext(ctx).Info("user event consumed")
}

// validate runs inside a nested child span. A missing id simulates a validation
// failure so the error-recording path can be shown.
func (r *Telemetry) validate(ctx context.Context, userID string) error {
	_, span := r.tracer.Start(ctx, "users.validate")
	defer span.End()

	if userID == "" {
		return errUserIDRequired
	}

	return nil
}
