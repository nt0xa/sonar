package netx

import (
	"context"
	"crypto/tls"
	"net"
)

type Server struct {
	Addr              string
	TLSConfig         *tls.Config
	NotifyStartedFunc func()
	ListenerWrapper   func(net.Listener) net.Listener // TODO: replace with handler
	Handler
}

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
}

type HandlerFunc func(context.Context, net.Conn)

func (f HandlerFunc) Handle(ctx context.Context, conn net.Conn) {
	f(ctx, conn)
}

func (s *Server) ListenAndServe() error {
	var (
		err      error
		listener net.Listener
	)

	if s.TLSConfig != nil {
		listener, err = tls.Listen("tcp", s.Addr, s.TLSConfig)
	} else {
		listener, err = net.Listen("tcp", s.Addr)
	}

	if err != nil {
		return err
	}

	defer func() {
		// TODO: logging
		_ = listener.Close()
	}()

	l := listener

	if s.ListenerWrapper != nil {
		l = s.ListenerWrapper(l)
	}

	if s.NotifyStartedFunc != nil {
		s.NotifyStartedFunc()
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go func() {
			// TODO: logging
			s.Handle(context.Background(), conn)
			_ = conn.Close()
		}()
	}
}
