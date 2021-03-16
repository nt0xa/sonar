package server

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/pkg/netx"
	"github.com/bi-zone/sonar/pkg/smtpx"
	"github.com/fatih/structs"
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

func SMTPSession(domain string, tlsConfig *tls.Config, notify func(*smtpx.Event)) func(net.Conn) *smtpx.Session {
	return func(conn net.Conn) *smtpx.Session {
		return smtpx.NewSession(conn, domain, tlsConfig, notify)
	}
}

func SMTPEvent(e *smtpx.Event) *models.Event {
	return &models.Event{
		Protocol:   models.ProtoSMTP,
		RW:         e.Log,
		Meta:       structs.Map(e.Msg),
		RemoteAddr: e.RemoteAddr.String(),
		ReceivedAt: e.ReceivedAt,
	}
}
