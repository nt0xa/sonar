package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/nt0xa/sonar/pkg/netx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server interface {
	ListenAndServe() error
}

// server contains parameters for running an HTTP server.
type server struct {
	// addr is address to listen on.
	addr string

	// handler is http.Handler to call.
	handler http.Handler

	// tlsConfig is a config for TLS.
	tlsConfig *tls.Config

	// h2c enables HTTP/2 cleartext support.
	h2c bool

	// If notifyStartedFunc is set it is called once the server has started listening.
	notifyStartedFunc func()

	server *http.Server
}

func New(addr string, h http.Handler, opts ...Option) Server {
	if h == nil {
		panic("httpx: handler must not be nil")
	}

	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	handler := http.Handler(h)

	// For non-TLS connections, wrap with h2c to support HTTP/2 cleartext.
	if options.h2c && options.tlsConfig == nil {
		handler = h2c.NewHandler(h, &http2.Server{})
	}

	srv := &http.Server{
		ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
			return saveConn(ctx, conn)
		},
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 12,
	}

	srv.SetKeepAlivesEnabled(false)

	// Register the HTTP/2 handler so TLS connections negotiating "h2" via
	// ALPN are served by the HTTP/2 stack.
	if err := http2.ConfigureServer(srv, nil); err != nil {
		panic("httpx: failed to configure http2: " + err.Error())
	}

	return &server{
		addr:              addr,
		handler:           h,
		notifyStartedFunc: options.notifyStartedFunc,
		tlsConfig:         options.tlsConfig,
		h2c:               options.h2c,
		server:            srv,
	}
}

// ListenAndServe starts the HTTP server on configured address.
func (srv *server) ListenAndServe() error {
	var (
		listener net.Listener
		err      error
	)

	if srv.tlsConfig != nil {
		// Clone the config so we don't mutate the caller's value, then
		// prepend "h2" and "http/1.1" to NextProtos so the TLS handshake
		// will advertise HTTP/2 support via ALPN.
		cfg := srv.tlsConfig.Clone()
		cfg.NextProtos = prependProtos(cfg.NextProtos, "h2", "http/1.1")
		listener, err = tls.Listen("tcp", srv.addr, cfg)
	} else {
		listener, err = net.Listen("tcp", srv.addr)
	}

	if err != nil {
		return err
	}

	if srv.notifyStartedFunc != nil {
		srv.notifyStartedFunc()
	}

	return srv.server.Serve(&netx.LoggingListener{
		Listener: listener,
	})
}

// prependProtos returns a new NextProtos slice with the given protocols at the
// front, deduplicating any that already exist.
func prependProtos(existing []string, protos ...string) []string {
	seen := make(map[string]bool, len(existing))
	for _, p := range existing {
		seen[p] = true
	}
	result := make([]string, 0, len(protos)+len(existing))
	for _, p := range protos {
		if !seen[p] {
			result = append(result, p)
			seen[p] = true
		}
	}
	return append(result, existing...)
}
