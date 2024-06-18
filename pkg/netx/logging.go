package netx

import (
	"bufio"
	"bytes"
	"io"
	"net"
)

// LoggingListener is net.Listener wrapper that returns wraps net.Conn with LoggingConn.
type LoggingListener struct {
	net.Listener
}

func (l *LoggingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return NewLoggingConn(conn), nil
}

// LoggingConn wraps net.Conn to save conversation log.
type LoggingConn struct {
	net.Conn

	// rw is ReaderWriter which reads/writes to connection and to log.
	rw *bufio.ReadWriter

	// RW is full conversation log.
	RW bytes.Buffer

	// R is read data.
	R bytes.Buffer

	// W is written data.
	W bytes.Buffer

	// onClose is called when connection is closed.
	// Must be used to get final conversation log.
	OnClose func()
}

// NewLoggingConn wraps net.Conn and adds logging.
func NewLoggingConn(conn net.Conn) *LoggingConn {
	c := &LoggingConn{
		Conn: conn,
	}

	// Reader that reads data from conn and write it to c.R and c.RW.
	r := io.TeeReader(io.TeeReader(conn, &c.R), &c.RW)

	// Writer that writes data to conn, c.W and c.RW.
	w := io.MultiWriter(conn, &c.W, &c.RW)

	// Create ReaderWriter for convenience.
	c.rw = bufio.NewReadWriter(bufio.NewReader(r), bufio.NewWriter(w))

	return c
}

// Write overwrites net.Conn Write method to be able to save data to log.
func (c *LoggingConn) Write(b []byte) (int, error) {
	n, err := c.rw.Write(b)
	if err != nil {
		return n, err
	}

	return n, c.rw.Flush()
}

// Read overwrites net.Conn Read method to be able to save data to log.
func (c *LoggingConn) Read(b []byte) (int, error) {
	return c.rw.Read(b)
}

// Close overwrites net.Conn Close method to call onClose function.
func (c *LoggingConn) Close() error {
	err := c.Conn.Close()

	if c.OnClose != nil {
		c.OnClose()
	}

	return err
}
