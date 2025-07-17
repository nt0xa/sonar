package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
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
func (d *noopTelemetry) TraceStart(
	ctx context.Context,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

var _ Telemetry = (*noopTelemetry)(nil)

// NewInt64Histogram implements Telemetry.
func (d *noopTelemetry) NewInt64Histogram(
	name string,
	unit string,
	description string,
	opts ...metric.HistogramOption,
) (Int64Histogram, error) {
	return new(noopMetric), nil
}

// NewInt64UpDownCounter implements Telemetry.
func (d *noopTelemetry) NewInt64UpDownCounter(
	name string,
	unit string,
	description string,
	opts ...metric.HistogramOption,
) (Int64UpDownCounter, error) {
	return new(noopMetric), nil
}

type noopMetric struct{}

// Add implements Int64UpDownCounter.
func (n *noopMetric) Add(ctx context.Context, incr int64, options ...metric.AddOption) {}

// Record implements Int64Histogram.
func (n *noopMetric) Record(ctx context.Context, incr int64, options ...metric.RecordOption) {}

var _ Int64Histogram = (*noopMetric)(nil)
var _ Int64UpDownCounter = (*noopMetric)(nil)
