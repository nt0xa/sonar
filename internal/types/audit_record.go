package types

import (
	"time"
)

type AuditRecord struct {
	ID           int64
	UUID         string
	CreatedAt    time.Time
	Action       string
	ResourceType string
	Source       string
	ActorID      *int64
	ActorName    string
	ActorMeta    map[string]any
	Resource     map[string]any
}
