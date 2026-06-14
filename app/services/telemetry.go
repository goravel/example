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

// errUserBlocked is a simulated failure used to show error recording on a span.
var errUserBlocked = errors.New("user is blocked")

type Telemetry interface {
	Process(ctx context.Context, userID string) error
	Publish(ctx context.Context) map[string]string
	Consume(headers map[string]string)
}

type TelemetryImpl struct{}

func NewTelemetryImpl() *TelemetryImpl {
	return &TelemetryImpl{}
}

// Process is a manually traced unit of work. It enriches the active (HTTP
// server) span, starts its own span with an attribute and an event, runs a
// nested child span, records the outcome on a counter and a histogram, tracks
// in-flight work with an up-down counter, and logs with the trace context.
func (r *TelemetryImpl) Process(ctx context.Context, userID string) error {
	start := time.Now()
	tracer := facades.Telemetry().Tracer(scopeName)
	meter := facades.Telemetry().Meter(scopeName)

	// Production code should create instruments once and reuse them, for example
	// on a long-lived service built at startup. They are resolved per call here
	// only to keep the example self-contained; the meter returns the same
	// instrument for a given name, so this stays correct and cheap.
	processed, _ := meter.Int64Counter("users.processed")
	duration, _ := meter.Float64Histogram("user.process.duration")
	inFlight, _ := meter.Int64UpDownCounter("users.in_flight")

	// Enrich the span already active on the incoming context (started by the
	// HTTP server middleware) before opening our own.
	trace.SpanFromContext(ctx).SetAttributes(telemetry.String("user.id", userID))

	ctx, span := tracer.Start(ctx, "users.process", telemetry.WithSpanKind(telemetry.SpanKindInternal))
	defer span.End()
	span.SetAttributes(telemetry.String("user.id", userID))
	span.AddEvent("user.validated")

	inFlight.Add(ctx, 1)
	defer inFlight.Add(ctx, -1)

	if err := r.validate(ctx, tracer, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(telemetry.CodeError, err.Error())
		processed.Add(ctx, 1, metric.WithAttributes(telemetry.String("result", "error")))
		duration.Record(ctx, time.Since(start).Seconds())
		facades.Log().WithContext(ctx).Error("user processing failed: ", err)

		return err
	}

	processed.Add(ctx, 1, metric.WithAttributes(telemetry.String("result", "ok")))
	duration.Record(ctx, time.Since(start).Seconds())
	facades.Log().WithContext(ctx).Info("user processed")

	return nil
}

// Publish injects the active trace context into message headers so a consumer
// on another transport can continue the same trace.
func (r *TelemetryImpl) Publish(ctx context.Context) map[string]string {
	headers := map[string]string{}
	facades.Telemetry().Propagator().Inject(ctx, telemetry.PropagationMapCarrier(headers))

	return headers
}

// Consume extracts the propagated context and starts a consumer span that
// continues the producer's trace across the boundary.
func (r *TelemetryImpl) Consume(headers map[string]string) {
	ctx := facades.Telemetry().Propagator().Extract(context.Background(), telemetry.PropagationMapCarrier(headers))

	_, span := facades.Telemetry().Tracer(scopeName).Start(ctx, "users.consume", telemetry.WithSpanKind(telemetry.SpanKindConsumer))
	defer span.End()

	facades.Log().WithContext(ctx).Info("user event consumed")
}

// validate runs inside a nested child span. An empty id simulates a failure so
// the error-recording path can be shown.
func (r *TelemetryImpl) validate(ctx context.Context, tracer trace.Tracer, userID string) error {
	_, span := tracer.Start(ctx, "users.validate")
	defer span.End()

	if userID == "" {
		return errUserBlocked
	}

	return nil
}
