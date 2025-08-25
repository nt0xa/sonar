package server

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/mail"
	"time"

	"github.com/fatih/structs"
	"go.opentelemetry.io/otel/attribute"
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
	log *slog.Logger,
	tel telemetry.Telemetry,
	tlsConfig *tls.Config,
	notify func(context.Context, *smtpx.Event),
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

func SMTPEvent(e *smtpx.Event) *models.Event {
	type Address struct {
		Name  string `structs:"name"`
		Email string `structs:"email"`
	}

	type Email struct {
		Subject string     `structs:"subject"`
		From    []Address  `structs:"from"`
		To      []Address  `structs:"to"`
		Cc      []Address  `structs:"cc"`
		Bcc     []Address  `structs:"bcc"`
		Date    *time.Time `structs:"date"`
		Text    string     `structs:"text"`
	}

	type Session struct {
		Helo     string   `structs:"helo"`
		Ehlo     string   `structs:"ehlo"`
		MailFrom string   `structs:"mailFrom"`
		RcptTo   []string `structs:"rcptTo"`
		Data     string   `structs:"data"`
	}

	type Meta struct {
		Session Session `structs:"session"`
		Email   Email   `structs:"email"`
		Secure  bool    `structs:"secure"`
	}

	addr := func(mm []*mail.Address) []Address {
		res := make([]Address, len(mm))
		for i, m := range mm {
			res[i] = Address{
				Name:  m.Name,
				Email: m.Address,
			}
		}
		return res
	}

	meta := &Meta{
		Session: Session{
			Helo:     e.Data.Helo,
			Ehlo:     e.Data.Ehlo,
			MailFrom: e.Data.MailFrom,
			RcptTo:   e.Data.RcptTo,
			Data:     e.Data.Data,
		},
		Email: Email{
			Subject: e.Email.Subject,
			From:    addr(e.Email.From),
			To:      addr(e.Email.To),
			Cc:      addr(e.Email.Cc),
			Bcc:     addr(e.Email.Bcc),
			Date:    e.Email.Date,
			Text:    e.Email.Text,
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
