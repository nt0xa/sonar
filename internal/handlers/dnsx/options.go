package dnsx

import (
	"net"

	"github.com/bi-zone/sonar/internal/handlers"
)

var defaultOptions = options{
	notifyStartedFunc: func() {},
	notifyRequestFunc: func(net.Addr, []byte, map[string]interface{}) {},
	subdomainPattern:  "[a-z0-9]{8}",
}

type options struct {
	notifyStartedFunc handlers.NotifyStartedFunc
	notifyRequestFunc handlers.NotifyRequestFunc
	subdomainPattern  string
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

func SubdomainPattern(p string) Option {
	return func(opts *options) {
		opts.subdomainPattern = p
	}
}
