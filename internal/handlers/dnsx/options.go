package dnsx

import (
	"net"

	"github.com/bi-zone/sonar/internal/handlers"
)

var defaultOptions = options{
	notifyStartedFunc: func() {},
	notifyRequestFunc: func(net.Addr, []byte, map[string]interface{}) {},
}

type options struct {
	notifyStartedFunc handlers.NotifyStartedFunc
	notifyRequestFunc handlers.NotifyRequestFunc
}

type Option func(*options)

func NotifyStartedFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}

func NotifyRequestFunc(f handlers.NotifyRequestFunc) Option {
	return func(opts *options) {
		opts.notifyRequestFunc = f
	}
}
