package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type DNSRecordsClear interface {
	DNSRecordsClear(context.Context, DNSRecordsClearInput) (DNSRecordsClearOutput, error)
}

type DNSRecordsClearInput struct {
	PayloadName string
	Name        string
}

func (in DNSRecordsClearInput) Validate() v.Problems {
	return v.Struct(
		v.String("payloadName", in.PayloadName).Required(),
	)
}

type DNSRecordsClearOutput = []DNSRecord
