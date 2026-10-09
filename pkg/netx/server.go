package netx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"time"
)

// maxAcceptDelay caps the backoff between failed accepts.
const maxAcceptDelay = time.Second

type Server struct {
	Addr              string
	TLSConfig         *tls.Config
	NotifyStartedFunc func()
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

	if s.NotifyStartedFunc != nil {
		s.NotifyStartedFunc()
	}

	return s.Serve(listener)
}

// Serve accepts connections on l until it is closed.
func (s *Server) Serve(l net.Listener) error {
	defer func() {
		// TODO: logging
		_ = l.Close()
	}()

	var delay time.Duration

	for {
		conn, err := l.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return err
			}

			// Back off so persistent errors like EMFILE don't spin the loop.
			delay = min(max(delay*2, 5*time.Millisecond), maxAcceptDelay)
			time.Sleep(delay)

			continue
		}

		delay = 0

		go func() {
			// TODO: logging
			s.Handle(context.Background(), conn)
			_ = conn.Close()
		}()
	}
}
