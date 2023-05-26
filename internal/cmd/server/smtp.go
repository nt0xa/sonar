package server

import (
	"net"
	"time"

	"github.com/fatih/structs"

	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/pkg/netx"
	"github.com/russtone/sonar/pkg/smtpx"
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
