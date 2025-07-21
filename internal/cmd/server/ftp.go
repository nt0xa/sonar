package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/fatih/structs"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/netx"
	"github.com/nt0xa/sonar/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func FTPListenerWrapper(maxBytes int64, idleTimeout time.Duration) func(net.Listener) net.Listener {
	return func(l net.Listener) net.Listener {
		return &netx.TimeoutListener{
			Listener: &netx.MaxBytesListener{
				Listener: l,
				MaxBytes: maxBytes,
			},
			IdleTimeout: idleTimeout,
		}
	}
}

func FTPHandler(
	domain string,
	tel telemetry.Telemetry,
	notify func(context.Context, *ftpx.Event),
) netx.Handler {
	return FTPTelemetry(
		ftpx.SessionHandler(
			ftpx.Msg{Greet: fmt.Sprintf("%s Server ready", domain)},
			notify,
		),
		tel,
	)
}

func FTPTelemetry(next netx.Handler, tel telemetry.Telemetry) netx.Handler {
	sessionDuration, err := tel.NewInt64Histogram(
		"ftp.session.duration",
		"ms",
		"FTP session duration",
	)
	if err != nil {
		panic(err)
	}

	counter, err := tel.NewInt64UpDownCounter(
		"ftp.sessions.inflight",
		"{count}",
		"Number of sessions currently being processed by the server",
	)
	if err != nil {
		panic(err)
	}

	return netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		start := time.Now()
		ctx, id := withEventID(ctx)

		ctx, span := tel.TraceStart(ctx, "ftp",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("event.id", id.String()),
			),
		)
		defer span.End()

		counter.Add(ctx, 1)
		next.Handle(ctx, conn)
		counter.Add(ctx, -1)
		sessionDuration.Record(ctx, time.Since(start).Milliseconds())
	})
}

func FTPEvent(e *ftpx.Event) *models.Event {
	type Meta struct {
		Session ftpx.Data `structs:"session"`
		Secure  bool      `structs:"secure"`
	}

	meta := &Meta{
		Session: e.Data,
		Secure:  e.Secure,
	}

	return &models.Event{
		Protocol:   models.ProtoFTP,
		R:          e.R,
		W:          e.W,
		RW:         e.RW,
		Meta:       structs.Map(meta),
		RemoteAddr: e.RemoteAddr.String(),
		ReceivedAt: e.ReceivedAt,
	}
}
