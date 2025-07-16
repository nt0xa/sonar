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
	notify func(*ftpx.Event),
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
	return netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		ctx, span := tel.TraceStart(ctx, "ftp",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(),
		)
		defer span.End()

		next.Handle(ctx, conn)
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
