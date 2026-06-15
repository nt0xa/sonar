package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type DNSRecordsList interface {
	DNSRecordsList(context.Context, DNSRecordsListInput) (DNSRecordsListOutput, error)
}

type DNSRecordsListInput struct {
	PayloadName string
}

func (in DNSRecordsListInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
	)
}

type DNSRecordsListOutput []DNSRecord
