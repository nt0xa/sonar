package smtpx

import (
	"crypto/tls"
	"net"
	"time"
)

var defaultOptions = options{
	idleTimeout:       time.Second * 30,
	sessionTimeout:    time.Second * 30,
	tlsConfig:         nil,
	startTLS:          false,
	notifyStartedFunc: func() {},
	notifyRequestFunc: func(net.Addr, []byte, map[string]interface{}) {},
}

type options struct {
	idleTimeout       time.Duration
	sessionTimeout    time.Duration
	tlsConfig         *tls.Config
	startTLS          bool
	notifyStartedFunc func()
	notifyRequestFunc func(net.Addr, []byte, map[string]interface{})
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

func StartTLS(b bool) Option {
	return func(opts *options) {
		opts.startTLS = b
	}
}

func NotifyStartedFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}

func NotifyRequestFunc(f func(net.Addr, []byte, map[string]interface{})) Option {
	return func(opts *options) {
		opts.notifyRequestFunc = f
	}
}
