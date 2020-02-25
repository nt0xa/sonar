package dns

import (
	"net"

	"github.com/bi-zone/sonar/pkg/server"
)

var defaultOptions = options{
	ttl:               1,
	notifyStartedFunc: func() {},
	notifyRequestFunc: func(net.Addr, []byte, map[string]interface{}) {},
}

type options struct {
	ttl               uint32
	notifyStartedFunc func()
	notifyRequestFunc server.NotifyRequestFunc
}

type Option func(*options)

func TTL(ttl uint32) Option {
	return func(opts *options) {
		opts.ttl = ttl
	}
}

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
