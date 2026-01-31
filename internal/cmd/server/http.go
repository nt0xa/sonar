package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/httpdb"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

// TODO: as parameters
const (
	httpHandlerTimeout = time.Second * 10
	httpMaxBodyBytes   = 1 << 20
)

func HTTPDefault(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	rnd, _ := utils.GenerateRandomString(8)
	_, _ = fmt.Fprintf(w, "<html><body>%s</body></html>", rnd)
}

func HTTPTelemetry(next http.Handler, tel telemetry.Telemetry) http.Handler {
	requestDuration, err := tel.NewInt64Histogram(
		"http.request.duration",
		"ms",
		"HTTP request duration",
	)
	if err != nil {
		panic(err)
	}

	counter, err := tel.NewInt64UpDownCounter(
		"http.requests.inflight",
		"{count}",
		"Number of requests currently being processed by the server",
	)
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		start := time.Now()

		ctx, id := withEventID(ctx)

		attrs := []attribute.KeyValue{
			attribute.String("event.id", id.String()),
		}

		if r.Method != "" {
			attrs = append(attrs, attribute.String("http.method", r.Method))
		} else {
			attrs = append(attrs, attribute.String("http.method", http.MethodGet))
		}

		if r.ContentLength >= 0 {
			attrs = append(attrs, attribute.Int64("http.request.content_length", r.ContentLength))
		}

		ctx, span := tel.TraceStart(ctx, "http",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attrs...,
			),
		)
		defer span.End()

		counter.Add(ctx, 1)

		next.ServeHTTP(w, r.WithContext(ctx))

		counter.Add(ctx, -1)
		requestDuration.Record(ctx, time.Since(start).Milliseconds())
	})
}

func HTTPHandler(
	db *database.DB,
	tel telemetry.Telemetry,
	origin string,
	notify httpx.NotifyFunc,
) http.Handler {
	return HTTPTelemetry(
		http.TimeoutHandler(
			httpx.BodyReaderHandler(
				httpx.MaxBytesHandler(
					httpx.NotifyHandler(
						notify,
						httpdb.Handler(
							&httpdb.Routes{DB: db, Origin: origin},
							http.HandlerFunc(HTTPDefault),
						),
					),
					httpMaxBodyBytes,
				),
				httpMaxBodyBytes,
			),
			httpHandlerTimeout,
			"timeout",
		),
		tel,
	)
}

func emitHTTP(events *EventsHandler) httpx.NotifyFunc {
	return func(
		ctx context.Context,
		remoteAddr net.Addr,
		receivedAt *time.Time,
		secure bool,
		read, written, combined []byte,
		meta *httpx.Meta,
	) {
		var proto string

		if secure {
			proto = database.ProtoHTTPS
		} else {
			proto = database.ProtoHTTP
		}

		events.Emit(ctx, &database.Event{
			Protocol: proto,
			R:        read,
			W:        written,
			RW:       combined,
			Meta: database.EventsMeta{
				HTTP:   meta,
				Secure: secure,
			},
			RemoteAddr: remoteAddr.String(),
			ReceivedAt: *receivedAt,
		})
	}
}
