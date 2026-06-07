package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type DNSRecordsList interface {
	DNSRecordsList(context.Context, DNSRecordsListInput) (DNSRecordsListOutput, error)
}

type DNSRecordsListInput struct {
	PayloadName string
}

func (in DNSRecordsListInput) Validate() v.Problems {
	return v.Struct(
		v.String("payloadName", in.PayloadName).Required(),
	)
}

type DNSRecordsListOutput = []DNSRecord
