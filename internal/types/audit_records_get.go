package types

import "context"

type AuditRecordsGet interface {
	AuditRecordsGet(context.Context, AuditRecordsGetInput) (*AuditRecordsGetOutput, error)
}

type AuditRecordsGetInput struct {
	ID int64
}

type AuditRecordsGetOutput = AuditRecord
