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
		server: &netx.Server{
			Addr:              addr,
			TLSConfig:         options.tlsConfig,
			NotifyStartedFunc: options.notifyStartedFunc,
			ListenerWrapper:   options.listenerWrapper,
			Handler:           handler,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}
