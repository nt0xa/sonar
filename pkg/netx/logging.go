package netx

import (
	"bufio"
	"net"
	"sync"
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

	// rw is ReaderWriter which reads/writes to connection.
	rw *bufio.ReadWriter

	// mu guards Data.
	mu sync.Mutex

	// Data is the ordered conversation log: each Read/Write call appends one
	// message, so reads and writes are interleaved in call order.
	Data [][]byte

	// onClose is called when connection is closed.
	// Must be used to get final conversation log.
	OnClose func()
}

// NewLoggingConn wraps net.Conn and adds logging.
func NewLoggingConn(conn net.Conn) *LoggingConn {
	c := &LoggingConn{
		Conn: conn,
	}

	c.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	return c
}

// append stores a copy of b as the next message in the conversation log.
func (c *LoggingConn) append(b []byte) {
	msg := make([]byte, len(b))
	copy(msg, b)

	c.mu.Lock()
	c.Data = append(c.Data, msg)
	c.mu.Unlock()
}

// Write overwrites net.Conn Write method to be able to save data to log.
func (c *LoggingConn) Write(b []byte) (int, error) {
	n, err := c.rw.Write(b)
	if err != nil {
		return n, err
	}

	if err := c.rw.Flush(); err != nil {
		return n, err
	}

	c.append(b[:n])

	return n, nil
}

// Read overwrites net.Conn Read method to be able to save data to log.
func (c *LoggingConn) Read(b []byte) (int, error) {
	n, err := c.rw.Read(b)
	if n > 0 {
		c.append(b[:n])
	}
	return n, err
}

// Close overwrites net.Conn Close method to call onClose function.
func (c *LoggingConn) Close() error {
	err := c.Conn.Close()

	if c.OnClose != nil {
		c.OnClose()
	}

	return err
}
