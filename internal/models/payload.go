package models

import (
	"time"
)

type Payload struct {
	ID              int64              `db:"id"`
	UserID          int64              `db:"user_id"`
	Subdomain       string             `db:"subdomain"`
	Name            string             `db:"name"`
	NotifyProtocols ProtoCategoryArray `db:"notify_protocols"`
	CreatedAt       time.Time          `db:"created_at"`
}
