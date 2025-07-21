package certmgr

import (
	"log/slog"
	"time"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
)

type options struct {
	keyType         certcrypto.KeyType
	caDirURL        string
	caInsecure      bool
	renewInterval   time.Duration
	renewThreshold  time.Duration
	notifyReadyFunc func()
	timeNow         func() time.Time
	log             *slog.Logger
}

var defaultOptions = options{
	keyType:         certcrypto.EC384,
	caDirURL:        lego.LEDirectoryProduction,
	caInsecure:      false,
	renewInterval:   12 * time.Hour,
	renewThreshold:  30 * 24 * time.Hour,
	notifyReadyFunc: func() {},
	timeNow:         time.Now,
	log:             slog.New(slog.DiscardHandler),
}

type Option func(*options)

func KeyType(keyType certcrypto.KeyType) Option {
	return func(opts *options) {
		opts.keyType = keyType
	}
}

func CADirURL(url string) Option {
	return func(opts *options) {
		opts.caDirURL = url
	}
}

func CAInsecure(insecure bool) Option {
	return func(opts *options) {
		opts.caInsecure = insecure
	}
}

func RenewInterval(d time.Duration) Option {
	return func(opts *options) {
		opts.renewInterval = d
	}
}

func RenewThreshold(d time.Duration) Option {
	return func(opts *options) {
		opts.renewThreshold = d
	}
}

func NotifyReadyFunc(f func()) Option {
	return func(opts *options) {
		opts.notifyReadyFunc = f
	}
}

func Logger(l *slog.Logger) Option {
	return func(opts *options) {
		opts.log = l
	}
}

func TestOnlyTimeNow(f func() time.Time) Option {
	return func(opts *options) {
		opts.timeNow = f
	}
}
