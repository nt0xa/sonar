package dnsx

var defaultOptions = options{
	notifyStartedFunc: func() {},
}

type options struct {
	notifyStartedFunc func()
}

type Option func(*options)

// WithNotifyStarted sets a function called once the server is listening.
func WithNotifyStarted(f func()) Option {
	return func(opts *options) {
		opts.notifyStartedFunc = f
	}
}
