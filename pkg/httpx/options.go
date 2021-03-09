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
