package modules

import (
	"context"

	"github.com/nt0xa/sonar/internal/database/models"
)

type Notification struct {
	User    *models.User
	Payload *models.Payload
	Event   *models.Event
}

// Notifier must be implemented by all modules, which are going to notify
// users about payload events.
type Notifier interface {
	Name() string

	// Notify is called every time payload event happens.
	Notify(context.Context, *Notification) error
}
