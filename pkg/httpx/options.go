package httpx

import "crypto/tls"

var defaultOptions = options{
	notifyStartedFunc: func() {},
	tlsConfig:         nil,
}

type options struct {
	notifyStartedFunc func()
	tlsConfig         *tls.Config
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
