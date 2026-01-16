package server

import (
	"context"
	"fmt"
	"log/slog"
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
	log *slog.Logger,
	tel telemetry.Telemetry,
	notify func(context.Context, *ftpx.Event),
) netx.Handler {
	return FTPTelemetry(
		ftpx.SessionHandler(
			ftpx.Msg{Greet: fmt.Sprintf("%s Server ready", domain)},
			log,
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
	ftpMeta := &models.FTPMeta{
		Session: models.FTPSession{
			User: e.Data.User,
			Pass: e.Data.Pass,
			Type: e.Data.Type,
			Pasv: e.Data.Pasv,
			Epsv: e.Data.Epsv,
			Port: e.Data.Port,
			Eprt: e.Data.Eprt,
			Retr: e.Data.Retr,
		},
		Secure: e.Secure,
	}

	return &models.Event{
		Protocol:   models.ProtoFTP,
		R:          e.R,
		W:          e.W,
		RW:         e.RW,
		Meta: models.Meta{
			FTPMeta: ftpMeta,
		},
		RemoteAddr: e.RemoteAddr.String(),
		ReceivedAt: e.ReceivedAt,
	}
}
