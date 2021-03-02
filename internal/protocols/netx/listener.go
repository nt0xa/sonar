package netx

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
	Addr              string
	Handler           Handler
	IdleTimeout       time.Duration
	SessionTimeout    time.Duration
	TLSConfig         *tls.Config
	NotifyStartedFunc func()

	listener net.Listener
}

func (l *Listener) Listen() error {

	var (
		err      error
		listener net.Listener
	)

	log.Printf("starting listener on %v", l.Addr)

	if l.TLSConfig != nil {
		listener, err = tls.Listen("tcp", l.Addr, l.TLSConfig)
	} else {
		listener, err = net.Listen("tcp", l.Addr)
	}

	if err != nil {
		return err
	}

	defer listener.Close()

	l.listener = listener

	if l.NotifyStartedFunc != nil {
		l.NotifyStartedFunc()
	}

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection %v", err)
			continue
		}

		log.Printf("accepted connection from %v", conn.RemoteAddr())

		c := &Conn{
			Conn:          conn,
			idleTimeout:   l.IdleTimeout,
			maxReadBuffer: 1 << 20,
		}

		if err := c.SetDeadline(time.Now().Add(l.IdleTimeout)); err != nil {
			return err
		}

		go func() {
			if err := l.handle(c); err != nil {
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
