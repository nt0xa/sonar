package service

import (
	"time"
)

//go:generate go-enum --ptr --names --values

// ENUM(create, update, delete)
type AuditAction string

// ENUM(payload, user, dns_record, http_route)
type AuditResourceType string

// ENUM(api, telegram, lark, slack)
type AuditSource string

type AuditRecord struct {
	ID           int64
	UUID         string
	CreatedAt    time.Time
	Action       AuditAction
	ResourceType AuditResourceType
	Source       AuditSource
	ActorID      *int64
	ActorName    string
	ActorMeta    map[string]any
	Resource     map[string]any
}
