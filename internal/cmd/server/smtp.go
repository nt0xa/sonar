package server

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/fatih/structs"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/database/models"
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
	tel telemetry.Telemetry,
	tlsConfig *tls.Config,
	notify func(*smtpx.Event),
) netx.Handler {
	return SMTPTelemetry(
		smtpx.SessionHandler(
			smtpx.Msg{Greet: domain, Ehlo: domain},
			tlsConfig,
			notify,
		),
		tel,
	)
}

func SMTPTelemetry(next netx.Handler, tel telemetry.Telemetry) netx.Handler {
	return netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		ctx, span := tel.TraceStart(ctx, "smtp",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(),
		)
		defer span.End()

		next.Handle(ctx, conn)
	})
}

func SMTPEvent(e *smtpx.Event) *models.Event {
	type Session struct {
		Helo     string   `structs:"helo"`
		Ehlo     string   `structs:"ehlo"`
		MailFrom string   `structs:"mailFrom"`
		RcptTo   []string `structs:"rcptTo"`
		Data     string   `structs:"data"`
	}

	type Meta struct {
		Session Session `structs:"session"`
		Secure  bool    `structs:"secure"`
	}

	meta := &Meta{
		Session: Session{
			Helo:     e.Data.Helo,
			Ehlo:     e.Data.Ehlo,
			MailFrom: e.Data.MailFrom,
			RcptTo:   e.Data.RcptTo,
			Data:     e.Data.Data,
		},
		Secure: e.Secure,
	}

	return &models.Event{
		Protocol:   models.ProtoSMTP,
		RW:         e.RW,
		R:          e.R,
		W:          e.W,
		Meta:       structs.Map(meta),
		RemoteAddr: e.RemoteAddr.String(),
		ReceivedAt: e.ReceivedAt,
	}
}
