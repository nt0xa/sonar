package smtp

import (
	"crypto/tls"
	"time"
)

var defaultOptions = options{
	idleTimeout:    time.Second * 30,
	sessionTimeout: time.Second * 30,
	tlsConfig:      nil,
	isTLS:          false,
}

type options struct {
	idleTimeout    time.Duration
	sessionTimeout time.Duration
	tlsConfig      *tls.Config
	isTLS          bool
}

type Option func(*options)

func IdleTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.idleTimeout = d
	}
}

func SessionTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.sessionTimeout = d
	}
}

func TLSConfig(c *tls.Config) Option {
	return func(opts *options) {
		opts.tlsConfig = c
	}
}

func IsTLS(b bool) Option {
	return func(opts *options) {
		opts.isTLS = b
	}
}
