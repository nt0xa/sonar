package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type AuditRecordsGet interface {
	AuditRecordsGet(context.Context, AuditRecordsGetInput) (*AuditRecordsGetOutput, error)
}

type AuditRecordsGetInput struct {
	ID int64
}

func (in AuditRecordsGetInput) Validate() v.Problems {
	return v.Struct(&in,
		v.Int(&in.ID, v.Required),
	)
}

type AuditRecordsGetOutput = AuditRecord
