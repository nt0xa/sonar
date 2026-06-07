package service

import (
	"context"
	"time"

	v "github.com/nt0xa/sonar/pkg/valid"
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

func (in AuditRecordsListInput) Validate() v.Problems {
	return v.Struct(&in,
		v.String(&in.ResourceType, v.Optional, v.In(AuditResourceTypeValues()...)),
		v.String(&in.Action, v.Optional, v.In(AuditActionValues()...)),
	)
}

type AuditRecordsListOutput = []AuditRecord
