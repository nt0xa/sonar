package netx

import (
	"io"
	"net"
)

type MaxBytesListener struct {
	net.Listener
	MaxBytes int64
}

func (l *MaxBytesListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	c := &MaxBytesConn{
		Conn: conn,
		r:    io.LimitReader(conn, l.MaxBytes),
	}

	return c, nil
}

type MaxBytesConn struct {
	net.Conn
	r io.Reader
}

func (c *MaxBytesConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}
