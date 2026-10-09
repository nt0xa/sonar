// Package netx provides a TCP server that runs a Handler per connection, plus
// connection middleware and wrappers.
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

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
}

type HandlerFunc func(context.Context, net.Conn)

func (f HandlerFunc) Handle(ctx context.Context, conn net.Conn) {
	f(ctx, conn)
}

type Server struct {
	addr          string
	handler       Handler
	tlsConfig     *tls.Config
	notifyStarted func()
}

// New creates a Server, panics on invalid arguments.
func New(addr string, handler Handler, opts ...Option) *Server {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	if handler == nil {
		panic("netx: handler must not be nil")
	}

	if options.notifyStarted == nil {
		panic("netx: notify started func must not be nil")
	}

	return &Server{
		addr:          addr,
		handler:       handler,
		tlsConfig:     options.tlsConfig,
		notifyStarted: options.notifyStarted,
	}
}

// ListenAndServe listens on the server address and serves connections.
func (s *Server) ListenAndServe() error {
	var (
		err      error
		listener net.Listener
	)

	if s.tlsConfig != nil {
		listener, err = tls.Listen("tcp", s.addr, s.tlsConfig)
	} else {
		listener, err = net.Listen("tcp", s.addr)
	}

	if err != nil {
		return err
	}

	s.notifyStarted()

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
			s.handler.Handle(context.Background(), conn)
			_ = conn.Close()
		}()
	}
}
