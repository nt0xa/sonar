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

func TLSConfig(c *tls.Config) Option {
	return func(opts *options) {
		opts.tlsConfig = c
	}
}

func NotifyStartedFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}
