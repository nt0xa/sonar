package server

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/bi-zone/sonar/pkg/netx"
	"github.com/bi-zone/sonar/pkg/smtpx"
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

func SMTPSession(domain string, tlsConfig *tls.Config, notify NotifyFunc) func(net.Conn) *smtpx.Session {
	return func(conn net.Conn) *smtpx.Session {
		return smtpx.NewSession(conn, domain, tlsConfig, func(data []byte, meta map[string]interface{}) {
			notify(conn.RemoteAddr(), data, meta)
		})
	}
}
