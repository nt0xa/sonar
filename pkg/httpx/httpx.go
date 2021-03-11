package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/bi-zone/sonar/pkg/netx"
)

type Server interface {
	ListenAndServe() error
}

// server contains parameters for running an DNS server.
type server struct {
	// addr is address to listen on.
	addr string

	// handler is http.Handler to call.
	handler http.Handler

	// tlsConfig is a config for TLS.
	tlsConfig *tls.Config

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

	return &server{
		addr:              addr,
		handler:           h,
		notifyStartedFunc: options.notifyStartedFunc,
		tlsConfig:         options.tlsConfig,
		server: &http.Server{
			ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
				return saveConn(ctx, conn)
			},
			Handler:        h,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 12,
		},
	}
}

// ListenAndServe starts the DNS server on configured address.
func (srv *server) ListenAndServe() error {
	var (
		listener net.Listener
		err      error
	)

	if srv.tlsConfig != nil {
		listener, err = tls.Listen("tcp", srv.addr, srv.tlsConfig)
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
