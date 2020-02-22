package http

import (
	"crypto/tls"
	"time"
)

var defaultOptions = options{
	idleTimeout:    time.Second * 5,
	sessionTimeout: time.Second * 5,
	tlsConfig:      nil,
}

type options struct {
	idleTimeout    time.Duration
	sessionTimeout time.Duration
	tlsConfig      *tls.Config
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
