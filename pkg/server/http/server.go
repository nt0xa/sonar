package http

import (
	"github.com/bi-zone/sonar/pkg/listener"
	"github.com/bi-zone/sonar/pkg/server"
)

type Server struct {
	addr        string
	options     *options
	handlerFunc server.HandlerFunc
}

func NewServer(addr string, handlerFunc server.HandlerFunc, opts ...Option) *Server {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	return &Server{
		addr:        addr,
		options:     &options,
		handlerFunc: handlerFunc,
	}
}

func (srv *Server) ListenAndServe() error {
	l := &listener.Listener{
		Addr:           srv.addr,
		Handler:        &Handler{srv.handlerFunc},
		IdleTimeout:    srv.options.idleTimeout,
		SessionTimeout: srv.options.sessionTimeout,
	}

	if srv.options.tlsConfig != nil {
		l.IsTLS = true
		l.TLSConfig = srv.options.tlsConfig
	}

	return l.Listen()
}
