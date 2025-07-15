package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type noopTelemetry struct{}

func NewNoop() Telemetry {
	return new(noopTelemetry)
}

// NewLogHandler implements Telemetry.
func (d *noopTelemetry) NewLogHandler(name string) slog.Handler {
	return slog.DiscardHandler
}

// Shutdown implements Telemetry.
func (d *noopTelemetry) Shutdown(ctx context.Context) error {
	return nil
}

// TraceStart implements Telemetry.
func (d *noopTelemetry) TraceStart(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

var _ Telemetry = (*noopTelemetry)(nil)
