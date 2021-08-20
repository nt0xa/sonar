package ftpx

import (
	"crypto/tls"
	"net"
	"time"
)

type options struct {
	sessionTimeout    time.Duration
	tlsConfig         *tls.Config
	notifyStartedFunc func()
	listenerWrapper   func(net.Listener) net.Listener
	onClose           func(*Event)
	messages          Msg
}

var defaultOptions = options{
	sessionTimeout:    time.Second * 30,
	tlsConfig:         nil,
	notifyStartedFunc: func() {},
	listenerWrapper:   func(l net.Listener) net.Listener { return l },
	onClose:           func(e *Event) {},
	messages:          Msg{""},
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

func ListenerWrapper(f func(net.Listener) net.Listener) Option {
	return func(opts *options) {
		opts.listenerWrapper = f
	}
}

func OnClose(f func(*Event)) Option {
	return func(opts *options) {
		opts.onClose = f
	}
}

func Messages(m Msg) Option {
	return func(opts *options) {
		opts.messages = m
	}
}
