package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

type Telemetry struct {
	lp *log.LoggerProvider
	mp *metric.MeterProvider
	tp *trace.TracerProvider
}

func New(ctx context.Context, name, version string) (*Telemetry, error) {
	res := newResource(name, version)

	lp, err := newLoggerProvider(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger provider: %w", err)
	}

	mp, err := newMetricProvider(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric provider: %w", err)
	}

	tp, err := newTracerProvider(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	return &Telemetry{
		lp: lp,
		mp: mp,
		tp: tp,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
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
