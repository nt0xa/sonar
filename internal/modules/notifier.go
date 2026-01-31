package modules

import (
	"context"

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

	// Notify is called every time payload event happens.
	Notify(context.Context, *Notification) error
}
