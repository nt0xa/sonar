package smtpx

import (
	"github.com/nt0xa/sonar/pkg/netx"
)

type Server struct {
	server *netx.Server
}

func New(addr string, handler netx.Handler, opts ...Option) *Server {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &Server{
		server: netx.New(addr, handler,
			netx.WithTLSConfig(options.tlsConfig),
			netx.WithNotifyStarted(options.notifyStartedFunc),
		),
	}
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}
