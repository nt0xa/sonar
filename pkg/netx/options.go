package netx

import "crypto/tls"

var defaultOptions = options{
	notifyStarted: func() {},
}

type options struct {
	tlsConfig     *tls.Config
	notifyStarted func()
}

type Option func(*options)

// WithTLSConfig makes the server accept TLS connections.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(opts *options) {
		opts.tlsConfig = cfg
	}
}

// WithNotifyStarted sets a function called once the server is listening.
func WithNotifyStarted(f func()) Option {
	return func(opts *options) {
		opts.notifyStarted = f
	}
}
