package netx

import (
	"context"
	"net"
	"time"
)

// TimeoutHandler fails reads and writes after idleTimeout of inactivity.
func TimeoutHandler(next Handler, idleTimeout time.Duration) Handler {
	return HandlerFunc(func(ctx context.Context, conn net.Conn) {
		c := &TimeoutConn{
			Conn:        conn,
			idleTimeout: idleTimeout,
		}

		if err := c.updateDeadline(); err != nil {
			return
		}

		next.Handle(ctx, c)
	})
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
	return c.SetDeadline(idleDeadline)
}
