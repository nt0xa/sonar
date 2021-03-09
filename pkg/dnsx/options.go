package dnsx

var defaultOptions = options{
	notifyStartedFunc: func() {},
}

type options struct {
	notifyStartedFunc func()
}

type Option func(*options)

func NotifyStartedFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}
