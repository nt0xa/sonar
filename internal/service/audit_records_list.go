package service

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
	ResourceType AuditResourceType
	Action       AuditAction
	From         *time.Time
	To           *time.Time
	Page         uint
	PerPage      uint
}

// NOTE: ResourceType/Action are value-type optional filters (empty = no filter).
// The valid package intentionally has no value-type optional, so they are not
// validated here; model them as pointers if validation is needed later.

type AuditRecordsListOutput = []AuditRecord
