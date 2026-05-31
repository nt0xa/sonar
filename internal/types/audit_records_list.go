package types

import (
	"context"
	"time"
)

type AuditRecordsList interface {
	AuditRecordsList(context.Context, AuditRecordsListInput) (AuditRecordsListOutput, error)
}

type AuditRecordsListInput struct {
	ActorID      *int64
	ActorName    string
	ResourceType string
	Action       string
	From         *time.Time
	To           *time.Time
	Page         uint
	PerPage      uint
}

type AuditRecordsListOutput []AuditRecord
