package ftpx

import "crypto/tls"

type options struct {
	tlsConfig         *tls.Config
	notifyStartedFunc func()
}

var defaultOptions = options{
	tlsConfig:         nil,
	notifyStartedFunc: func() {},
}

type Option func(*options)

// WithTLSConfig makes the server accept TLS connections.
func WithTLSConfig(c *tls.Config) Option {
	return func(opts *options) {
		opts.tlsConfig = c
	}
}

// WithNotifyStarted sets a function called once the server is listening.
func WithNotifyStarted(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}
