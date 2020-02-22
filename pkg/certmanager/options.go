package certmanager

import (
	"github.com/go-acme/lego/v3/certcrypto"
)

type options struct {
	days    int
	keyType certcrypto.KeyType
}

var defaultOptions = options{
	days:    30,
	keyType: certcrypto.EC384,
}

type Option func(*options)

func Days(days int) Option {
	return func(opts *options) {
		opts.days = days
	}
}

func KeyType(keyType certcrypto.KeyType) Option {
	return func(opts *options) {
		opts.keyType = keyType
	}
}
