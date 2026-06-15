package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type AuditRecordsGet interface {
	AuditRecordsGet(context.Context, AuditRecordsGetInput) (*AuditRecordsGetOutput, error)
}

type AuditRecordsGetInput struct {
	ID int64
}

func (in AuditRecordsGetInput) Validate() valid.Problems {
	return valid.Validate(
		valid.Number("id", in.ID, valid.Required),
	)
}

type AuditRecordsGetOutput AuditRecord
