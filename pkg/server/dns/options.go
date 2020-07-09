package dns

import (
	"net"

	"github.com/bi-zone/sonar/pkg/server"
)

var defaultOptions = options{
	notifyStartedFunc: func() {},
	notifyRequestFunc: func(net.Addr, []byte, map[string]interface{}) {},
}

type options struct {
	notifyStartedFunc func()
	notifyRequestFunc server.NotifyRequestFunc
}

type Option func(*options)

func NotifyStartedFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}

func NotifyRequestFunc(f server.NotifyRequestFunc) Option {
	return func(opts *options) {
		opts.notifyRequestFunc = f
	}
}
