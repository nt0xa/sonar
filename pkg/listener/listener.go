package listener

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
)

type Handler interface {
	Handle(context.Context, *Conn) error
}

type Conn struct {
	net.Conn
	idleTimeout   time.Duration
	maxReadBuffer int64
}

func (c *Conn) Write(p []byte) (int, error) {
	if err := c.updateDeadline(); err != nil {
		return 0, err
	}
	return c.Conn.Write(p)
}

func (c *Conn) Read(b []byte) (int, error) {
	if err := c.updateDeadline(); err != nil {
		return 0, err
	}
	return io.LimitReader(c.Conn, c.maxReadBuffer).Read(b)
}

func (c *Conn) updateDeadline() error {
	idleDeadline := time.Now().Add(c.idleTimeout)
	return c.Conn.SetDeadline(idleDeadline)
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}

type Listener struct {
	Addr           string
	Handler        Handler
	IdleTimeout    time.Duration
	SessionTimeout time.Duration
	TLSConfig      *tls.Config
	IsTLS          bool

	listener net.Listener
}

func (s *Listener) Listen() error {

	var (
		err      error
		listener net.Listener
	)

	log.Printf("starting listener on %v", s.Addr)

	if s.IsTLS && s.TLSConfig != nil {
		listener, err = tls.Listen("tcp", s.Addr, s.TLSConfig)
	} else {
		listener, err = net.Listen("tcp", s.Addr)
	}

	if err != nil {
		return err
	}

	defer listener.Close()

	s.listener = listener

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection %v", err)
			continue
		}

		log.Printf("accepted connection from %v", conn.RemoteAddr())

		c := &Conn{
			Conn:          conn,
			idleTimeout:   s.IdleTimeout,
			maxReadBuffer: 1 << 20,
		}

		if err := c.SetDeadline(time.Now().Add(s.IdleTimeout)); err != nil {
			return err
		}

		go func() {
			if err := s.handle(c); err != nil {
				log.Printf("fail to handle connection %v (%v)", err, c.RemoteAddr())
			}
		}()
	}
}

func (l *Listener) handle(conn *Conn) error {
	defer func() {
		log.Printf("closing connection from %v", conn.RemoteAddr())

		if err := conn.Close(); err != nil {
			log.Printf("error while closing connection (%v)", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), l.SessionTimeout)
	defer cancel()

	return l.Handler.Handle(ctx, conn)

}
