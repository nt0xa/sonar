package netx

import (
	"context"
	"net"
)

// RateLimitHandler drops connections for which allow returns false.
func RateLimitHandler(next Handler, allow func(net.Addr) bool) Handler {
	return HandlerFunc(func(ctx context.Context, conn net.Conn) {
		if !allow(conn.RemoteAddr()) {
			return
		}

		next.Handle(ctx, conn)
	})
}

// RateLimitListener closes accepted connections for which Allow returns false.
type RateLimitListener struct {
	net.Listener
	Allow func(net.Addr) bool
}

func (l *RateLimitListener) Accept() (net.Conn, error) {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}

		if l.Allow(conn.RemoteAddr()) {
			return conn, nil
		}

		_ = conn.Close()
	}
}
