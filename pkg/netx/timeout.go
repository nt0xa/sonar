package netx

import (
	"net"
	"time"
)

type TimeoutListener struct {
	net.Listener
	IdleTimeout time.Duration
}

func (l *TimeoutListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	c := &TimeoutConn{
		Conn:        conn,
		idleTimeout: l.IdleTimeout,
	}

	if err := c.SetDeadline(time.Now().Add(l.IdleTimeout)); err != nil {
		return nil, err
	}

	return c, nil
}

type TimeoutConn struct {
	net.Conn
	idleTimeout time.Duration
}

func (c *TimeoutConn) Write(b []byte) (int, error) {
	if err := c.updateDeadline(); err != nil {
		return 0, err
	}

	return c.Conn.Write(b)
}

func (c *TimeoutConn) Read(b []byte) (int, error) {
	if err := c.updateDeadline(); err != nil {
		return 0, err
	}

	return c.Conn.Read(b)
}

func (c *TimeoutConn) updateDeadline() error {
	idleDeadline := time.Now().Add(c.idleTimeout)
	return c.Conn.SetDeadline(idleDeadline)
}
