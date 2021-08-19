package smtpx

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/bi-zone/sonar/pkg/netx"
)

type Server struct {
	server *netx.Server
}

func New(addr string, opts ...Option) *Server {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	var tlsConfig *tls.Config

	// Start TLS server only if secure flag is set.
	if options.secure {
		tlsConfig = options.tlsConfig
	}

	return &Server{
		server: &netx.Server{
			Addr:              addr,
			TLSConfig:         tlsConfig,
			NotifyStartedFunc: options.notifyStartedFunc,
			ListenerWrapper:   options.listenerWrapper,
			ConnectionHandler: func(conn net.Conn) error {
				ctx, cancel := context.WithTimeout(context.Background(), options.sessionTimeout)
				defer cancel()
				return handleConn(ctx, conn, options)
			},
		},
	}

}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}
