package smtp

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/bi-zone/sonar/pkg/listener"
	"github.com/bi-zone/sonar/pkg/server"
)

type HandlerFunc func(net.Addr, string, []byte)

type Handler struct {
	domain      string
	tlsConfig   *tls.Config
	handlerFunc server.HandlerFunc
}

func (h *Handler) Handle(ctx context.Context, conn *listener.Conn) error {
	s := newSession(conn, h.domain, h.tlsConfig)

	s.onClose(func(data Data, log []byte) {
		if h.handlerFunc != nil {
			h.handlerFunc(conn.RemoteAddr(), "SMTP", log)
		}
	})

	return s.start(ctx)
}
