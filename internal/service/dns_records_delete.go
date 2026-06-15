package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type DNSRecordsDelete interface {
	DNSRecordsDelete(context.Context, DNSRecordsDeleteInput) (*DNSRecordsDeleteOutput, error)
}

type DNSRecordsDeleteInput struct {
	PayloadName string
	Index       int64
}

func (in DNSRecordsDeleteInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
		valid.Number("index", in.Index, valid.Required),
	)
}

type DNSRecordsDeleteOutput DNSRecord
