package smtp

import (
	"errors"

	"github.com/bi-zone/sonar/pkg/listener"
	"github.com/bi-zone/sonar/pkg/server"
)

type Server struct {
	addr        string
	domain      string
	options     *options
	handlerFunc server.HandlerFunc
}

func NewServer(addr, domain string, handlerFunc server.HandlerFunc, opts ...Option) *Server {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	return &Server{
		addr:        addr,
		domain:      domain,
		options:     &options,
		handlerFunc: handlerFunc,
	}
}

func (srv *Server) ListenAndServe() error {
	l := &listener.Listener{
		Addr: srv.addr,
		Handler: &Handler{
			handlerFunc: srv.handlerFunc,
			domain:      srv.domain,
			tlsConfig:   srv.options.tlsConfig,
		},
		IdleTimeout:    srv.options.idleTimeout,
		SessionTimeout: srv.options.sessionTimeout,
	}

	if srv.options.isTLS && srv.options.tlsConfig == nil {
		return errors.New("invalid TLS config")
	}

	if srv.options.tlsConfig != nil {
		l.IsTLS = srv.options.isTLS
		l.TLSConfig = srv.options.tlsConfig
	}

	return l.Listen()
}
