package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type DNSRecordsDelete interface {
	DNSRecordsDelete(context.Context, DNSRecordsDeleteInput) (*DNSRecordsDeleteOutput, error)
}

type DNSRecordsDeleteInput struct {
	PayloadName string
	Index       int64
}

func (in DNSRecordsDeleteInput) Validate() v.Problems {
	return v.Validate(
		v.String("payloadName", in.PayloadName, v.Required),
		v.Number("index", in.Index, v.Required),
	)
}

type DNSRecordsDeleteOutput DNSRecord
