package httpx

import (
	"crypto/tls"
	"net"
)

var defaultOptions = options{
	notifyStartedFunc: func() {},
	tlsConfig:         nil,
	listenerWrapper:   func(l net.Listener) net.Listener { return l },
}

type options struct {
	notifyStartedFunc func()
	tlsConfig         *tls.Config
	listenerWrapper   func(net.Listener) net.Listener
}

type Option func(*options)

// WithNotifyStarted sets a function called once the server is listening.
func WithNotifyStarted(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}

// WithTLSConfig makes the server accept TLS connections.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(opts *options) {
		opts.tlsConfig = cfg
	}
}

// WithListenerWrapper wraps the listener, e.g. to reject connections before the TLS handshake.
func WithListenerWrapper(f func(net.Listener) net.Listener) Option {
	return func(opts *options) {
		opts.listenerWrapper = f
	}
}
