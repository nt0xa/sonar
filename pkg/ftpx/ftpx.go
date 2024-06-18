package ftpx

import (
	"context"
	"net"

	"github.com/nt0xa/sonar/pkg/netx"
)

type Server struct {
	server *netx.Server
}

func New(addr string, opts ...Option) *Server {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &Server{
		server: &netx.Server{
			Addr:              addr,
			TLSConfig:         options.tlsConfig,
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
