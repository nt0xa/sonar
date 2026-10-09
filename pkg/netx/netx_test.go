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
		srv := &netx.Server{Handler: netx.HandlerFunc(func(context.Context, net.Conn) {})}

		errc := make(chan error)
		go func() { errc <- srv.Serve(l) }()

		time.Sleep(time.Second)

		// Without backoff this would be millions.
		assert.Less(t, l.accepts.Load(), int64(20))

		_ = l.Close()
		assert.ErrorIs(t, <-errc, net.ErrClosed)
	})
}
