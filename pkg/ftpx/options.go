package ftpx

import (
	"crypto/tls"
	"net"
)

type options struct {
	tlsConfig         *tls.Config
	notifyStartedFunc func()
	listenerWrapper   func(net.Listener) net.Listener
}

var defaultOptions = options{
	tlsConfig:         nil,
	notifyStartedFunc: func() {},
	listenerWrapper:   func(l net.Listener) net.Listener { return l },
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

func ListenerWrapper(f func(net.Listener) net.Listener) Option {
	return func(opts *options) {
		opts.listenerWrapper = f
	}
}
