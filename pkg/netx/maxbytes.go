package netx

import (
	"context"
	"io"
	"net"
)

// MaxBytesHandler limits the total number of bytes read from the connection.
func MaxBytesHandler(next Handler, maxBytes int64) Handler {
	return HandlerFunc(func(ctx context.Context, conn net.Conn) {
		next.Handle(ctx, &MaxBytesConn{
			Conn: conn,
			r:    io.LimitReader(conn, maxBytes),
		})
	})
}

type MaxBytesConn struct {
	net.Conn
	r io.Reader
}

func (c *MaxBytesConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}
