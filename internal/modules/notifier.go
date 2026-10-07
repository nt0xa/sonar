package modules

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/nt0xa/sonar/internal/database"
)

type Notification struct {
	User    *database.User
	Payload *database.Payload
	Event   *database.Event
}

// Notifier must be implemented by all modules, which are going to notify
// users about payload events.
type Notifier interface {
	Name() string

	// Notify is called for a payload event that is sent in full.
	Notify(context.Context, *Notification) error

	// NotifyBatch is called with events batched over a window, all for the same user and payload.
	NotifyBatch(context.Context, []*Notification) error

	// RateLimit returns how fast Notify and NotifyBatch may be called, per messenger API limits.
	RateLimit() (rate.Limit, int)
}
