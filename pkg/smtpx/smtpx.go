package smtpx

import (
	"context"
	"net"

	"github.com/bi-zone/sonar/pkg/netx"
)

type Server interface {
	ListenAndServe() error
}

type ListenerWrapFunc func(net.Listener) net.Listener

type NewSessionFunc func(net.Conn) *Session

type server struct {
	server *netx.Server
}

func New(addr string, wrap ListenerWrapFunc, sess NewSessionFunc, opts ...Option) Server {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &server{
		server: &netx.Server{
			Addr:              addr,
			TLSConfig:         options.tlsConfig,
			NotifyStartedFunc: options.notifyStartedFunc,
			ListenerWrapper:   wrap,
			ConnectionHandler: func(conn net.Conn) error {
				ctx, cancel := context.WithTimeout(context.Background(), options.sessionTimeout)
				defer cancel()
				return sess(conn).start(ctx)
			},
		},
	}
}

func (s *server) ListenAndServe() error {
	return s.server.ListenAndServe()
}
