package server

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/pkg/netx"
	"github.com/nt0xa/sonar/pkg/smtpx"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

func SMTPListenerWrapper(maxBytes int64, idleTimeout time.Duration) func(net.Listener) net.Listener {
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

func SMTPHandler(
	domain string,
	log *slog.Logger,
	tel telemetry.Telemetry,
	tlsConfig *tls.Config,
	notify smtpx.OnCloseFunc,
) netx.Handler {
	return SMTPTelemetry(
		smtpx.SessionHandler(
			smtpx.Msg{Greet: domain, Ehlo: domain},
			log,
			tlsConfig,
			notify,
		),
		tel,
	)
}

func SMTPTelemetry(next netx.Handler, tel telemetry.Telemetry) netx.Handler {
	sessionDuration, err := tel.NewInt64Histogram(
		"smtp.session.duration",
		"ms",
		"SMTP session duration",
	)
	if err != nil {
		panic(err)
	}

	counter, err := tel.NewInt64UpDownCounter(
		"smtp.sessions.inflight",
		"{count}",
		"Number of sessions currently being processed by the server",
	)
	if err != nil {
		panic(err)
	}

	return netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		start := time.Now()
		ctx, id := withEventID(ctx)

		ctx, span := tel.TraceStart(ctx, "smtp",
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

func emitSMTP(events *EventsHandler) smtpx.OnCloseFunc {
	return func(
		ctx context.Context,
		remoteAddr net.Addr,
		receivedAt *time.Time,
		secure bool,
		read, written, combined []byte,
		meta *smtpx.Meta,
	) {
		events.Emit(ctx, &database.Event{
			Protocol: database.ProtoSMTP,
			RW:       combined,
			R:        read,
			W:        written,
			Meta: database.EventsMeta{
				SMTP:   meta,
				Secure: secure,
			},
			RemoteAddr: remoteAddr.String(),
			ReceivedAt: *receivedAt,
		})
	}
}
