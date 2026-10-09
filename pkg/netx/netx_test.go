package netx_test

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/netx"
)

func Test_Panics(t *testing.T) {
	h := netx.HandlerFunc(func(context.Context, net.Conn) {})

	require.Panics(t, func() { netx.New("", nil) })
	require.Panics(t, func() { netx.New("", h, netx.WithNotifyStarted(nil)) })
}

func Test_TimeoutHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		srv, cli := net.Pipe()
		defer func() { _ = cli.Close() }()

		go func() {
			time.Sleep(900 * time.Millisecond)
			_, _ = cli.Write([]byte("a"))
		}()

		start := time.Now()

		netx.TimeoutHandler(netx.HandlerFunc(func(_ context.Context, conn net.Conn) {
			buf := make([]byte, 1)

			// Data arrives before the idle timeout and extends the deadline.
			_, err := conn.Read(buf)
			require.NoError(t, err)

			_, err = conn.Read(buf)
			assert.ErrorIs(t, err, os.ErrDeadlineExceeded)
			assert.Equal(t, 1900*time.Millisecond, time.Since(start))
		}), time.Second).Handle(t.Context(), srv)
	})
}

func Test_MaxBytesHandler(t *testing.T) {
	srv, cli := net.Pipe()

	go func() {
		_, _ = cli.Write([]byte("hello world"))
		_ = cli.Close()
	}()

	netx.MaxBytesHandler(netx.HandlerFunc(func(_ context.Context, conn net.Conn) {
		data, err := io.ReadAll(conn)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))
	}), 5).Handle(t.Context(), srv)

	// Unblocks the writer.
	_ = srv.Close()
}

// errListener fails every Accept until closed.
type errListener struct {
	accepts   atomic.Int64
	closed    chan struct{}
	closeOnce sync.Once
}

func (l *errListener) Accept() (net.Conn, error) {
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	default:
	}

	// Stop a spinning loop so the test fails instead of hanging.
	if l.accepts.Add(1) > 1000 {
		_ = l.Close()
	}

	return nil, errors.New("too many open files")
}

func (l *errListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	return nil
}

func (l *errListener) Addr() net.Addr { return nil }

func Test_ServeBacksOffOnAcceptErrorsAndStopsWhenClosed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := &errListener{closed: make(chan struct{})}
		srv := netx.New("", netx.HandlerFunc(func(context.Context, net.Conn) {}))

		errc := make(chan error)
		go func() { errc <- srv.Serve(l) }()

		time.Sleep(time.Second)

		// Without backoff this would be millions.
		assert.Less(t, l.accepts.Load(), int64(20))

		_ = l.Close()
		assert.ErrorIs(t, <-errc, net.ErrClosed)
	})
}

func Test_LoggingHandler(t *testing.T) {
	srv, cli := net.Pipe()
	defer func() { _ = cli.Close() }()

	go func() {
		_, _ = cli.Write([]byte("ping"))
		_, _ = io.ReadFull(cli, make([]byte, 4))
	}()

	assert.Nil(t, netx.LoggingConnFromContext(t.Context()))

	netx.LoggingHandler(netx.HandlerFunc(func(ctx context.Context, conn net.Conn) {
		rec := netx.LoggingConnFromContext(ctx)
		require.NotNil(t, rec)
		assert.Same(t, rec, conn)

		_, err := io.ReadFull(conn, make([]byte, 4))
		require.NoError(t, err)

		_, err = conn.Write([]byte("pong"))
		require.NoError(t, err)

		assert.Equal(t, [][]byte{[]byte("ping"), []byte("pong")}, rec.Data)
	})).Handle(t.Context(), srv)
}

func Test_LoggingConnUpgradeKeepsLog(t *testing.T) {
	srv1, cli1 := net.Pipe()
	srv2, cli2 := net.Pipe()
	defer func() { _ = cli1.Close(); _ = cli2.Close() }()

	go func() {
		_, _ = cli1.Write([]byte("one"))
		_ = cli1.Close()
	}()
	go func() { _, _ = cli2.Write([]byte("two")) }()

	c := netx.NewLoggingConn(srv1)

	_, err := io.ReadFull(c, make([]byte, 3))
	require.NoError(t, err)

	c.Upgrade(srv2)

	_, err = io.ReadFull(c, make([]byte, 3))
	require.NoError(t, err)

	assert.Same(t, srv2, c.Conn)
	assert.Equal(t, [][]byte{[]byte("one"), []byte("two")}, c.Data)
}

func Test_RateLimitHandler(t *testing.T) {
	var handled []string

	h := netx.RateLimitHandler(netx.HandlerFunc(func(_ context.Context, conn net.Conn) {
		handled = append(handled, conn.RemoteAddr().String())
	}), func(addr net.Addr) bool {
		return addr.String() == "allowed"
	})

	h.Handle(t.Context(), &addrConn{addr: "allowed"})
	h.Handle(t.Context(), &addrConn{addr: "rejected"})

	assert.Equal(t, []string{"allowed"}, handled)
}

// addrConn is a net.Conn with a fixed remote address.
type addrConn struct {
	net.Conn
	addr string
}

func (c *addrConn) RemoteAddr() net.Addr { return &net.UnixAddr{Name: c.addr} }

func Test_RateLimitListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	var calls atomic.Int64

	// Reject every other connection, starting with the first.
	l := &netx.RateLimitListener{
		Listener: ln,
		Allow: func(net.Addr) bool {
			return calls.Add(1)%2 == 0
		},
	}
	defer func() { _ = l.Close() }()

	rejected, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	defer func() { _ = rejected.Close() }()

	allowed, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	defer func() { _ = allowed.Close() }()

	conn, err := l.Accept()
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	assert.Equal(t, allowed.LocalAddr().String(), conn.RemoteAddr().String())
	assert.EqualValues(t, 2, calls.Load())

	// The rejected connection is closed by the server.
	require.NoError(t, rejected.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, err = rejected.Read(make([]byte, 1))
	assert.ErrorIs(t, err, io.EOF)
}
