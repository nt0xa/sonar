package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type telemetry struct {
	lp     *log.LoggerProvider
	mp     *metric.MeterProvider
	tp     *trace.TracerProvider
	meter  otelmetric.Meter
	tracer oteltrace.Tracer
}

type Telemetry interface {
	TraceStart(ctx context.Context, name string) (context.Context, oteltrace.Span)
	NewLogger(name string) *slog.Logger
	Shutdown(ctx context.Context) error
}

func New(ctx context.Context, name, version string) (Telemetry, error) {
	res := newResource(name, version)

	lp, err := newLoggerProvider(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger provider: %w", err)
	}

	mp, err := newMetricProvider(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric provider: %w", err)
	}
	meter := mp.Meter(name)

	tp, err := newTracerProvider(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}
	tracer := tp.Tracer(name)

	return &telemetry{
		lp:     lp,
		mp:     mp,
		tp:     tp,
		meter:  meter,
		tracer: tracer,
	}, nil
}

func (t *telemetry) Shutdown(ctx context.Context) error {
	if err := t.lp.Shutdown(ctx); err != nil {
		return err
	}

	if err := t.mp.Shutdown(ctx); err != nil {
		return err
	}

	if err := t.tp.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func (t *telemetry) TraceStart(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	//nolint: spancheck // this is a wrapper around the otel tracer
	return t.tracer.Start(ctx, name)
}

func (t *telemetry) NewLogger(name string) *slog.Logger {
	return otelslog.NewLogger(name, otelslog.WithLoggerProvider(t.lp))
}

func newResource(name, version string) *resource.Resource {
	hostname, _ := os.Hostname()

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(name),
		semconv.ServiceVersion(version),
		semconv.HostName(hostname),
	)
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) (*log.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	processor := log.NewBatchProcessor(exporter)

	lp := log.NewLoggerProvider(
		log.WithProcessor(processor),
		log.WithResource(res),
	)

	return lp, nil
}

func newMetricProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return mp, nil
}

func newTracerProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp, nil
}
