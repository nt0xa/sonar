package dns

var defaultOptions = options{
	ttl:               1,
	notifyStartedFunc: func() {},
}

type options struct {
	ttl               uint32
	notifyStartedFunc func()
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
