package models

import (
	"time"

	"github.com/lib/pq"
)

type Payload struct {
	ID        int64          `db:"id"         json:"-"`
	UserID    int64          `db:"user_id"    json:"-"`
	Subdomain string         `db:"subdomain"  json:"subdomain"`
	Name      string         `db:"name"       json:"name"`
	Handlers  pq.StringArray `db:"handlers"   json:"handlers,omitempty"`
	CreatedAt time.Time      `db:"created_at" json:"createdAt"`
}
