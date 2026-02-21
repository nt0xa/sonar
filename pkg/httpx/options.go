package httpx

import "crypto/tls"

var defaultOptions = options{
	notifyStartedFunc: func() {},
	tlsConfig:         nil,
	h2c:               false,
}

type options struct {
	notifyStartedFunc func()
	tlsConfig         *tls.Config
	h2c               bool
}

type Option func(*options)

func NotifyStartedFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}

func TLSConfig(cfg *tls.Config) Option {
	return func(opts *options) {
		opts.tlsConfig = cfg
	}
}

// H2C enables HTTP/2 cleartext (h2c) support for non-TLS servers.
// Clients can then upgrade to HTTP/2 via the Upgrade header or use
// HTTP/2 prior knowledge mode.
func H2C() Option {
	return func(opts *options) {
		opts.h2c = true
	}
}
