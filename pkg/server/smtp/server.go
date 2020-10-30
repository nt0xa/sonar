package smtp

import (
	"context"

	"github.com/bi-zone/sonar/pkg/listener"
)

type Server struct {
	addr    string
	domain  string
	options *options
}

func New(addr, domain string, opts ...Option) *Server {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	return &Server{
		addr:    addr,
		domain:  domain,
		options: &options,
	}
}

func (s *Server) SetOption(opt Option) {
	opt(s.options)
}

func (s *Server) ListenAndServe() error {
	l := &listener.Listener{
		Addr:              s.addr,
		Handler:           s,
		IdleTimeout:       s.options.idleTimeout,
		SessionTimeout:    s.options.sessionTimeout,
		NotifyStartedFunc: s.options.notifyStartedFunc,
	}

	// With STARTTLS TLS connection is established during the session
	// after "STARTTLS" command, so we don't need TLS listener here
	if s.options.tlsConfig != nil && !s.options.startTLS {
		l.TLSConfig = s.options.tlsConfig
	}

	return l.Listen()
}

func (s *Server) Handle(ctx context.Context, conn *listener.Conn) error {
	sess := newSession(conn, s.domain, s.options.tlsConfig)

	sess.onClose(func(log []byte, meta map[string]interface{}) {
		if s.options.notifyRequestFunc != nil {
			s.options.notifyRequestFunc(conn.RemoteAddr(), log, meta)
		}
	})

	return sess.start(ctx)
}
