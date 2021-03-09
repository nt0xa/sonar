package smtpx

import (
	"crypto/tls"
	"time"
)

var defaultOptions = options{
	sessionTimeout:    time.Second * 30,
	tlsConfig:         nil,
	maxSessionBytes:   1 << 20,
	notifyStartedFunc: func() {},
}

type options struct {
	idleTimeout       time.Duration
	sessionTimeout    time.Duration
	tlsConfig         *tls.Config
	notifyStartedFunc func()
	maxSessionBytes   int64
}

type Option func(*options)

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

func NotifyStartedFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}

func MaxSessionBytes(n int64) Option {
	return func(opts *options) {
		opts.maxSessionBytes = n
	}
}
