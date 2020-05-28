package storage

import (
	"os"
)

type options struct {
	filePerm os.FileMode
}

var defaultOptions = options{
	filePerm: 0600,
}

type Option func(*options)

func FilePerm(perm os.FileMode) Option {
	return func(opts *options) {
		opts.filePerm = perm
	}
}
