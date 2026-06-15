package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type DNSRecordsClear interface {
	DNSRecordsClear(context.Context, DNSRecordsClearInput) (DNSRecordsClearOutput, error)
}

type DNSRecordsClearInput struct {
	PayloadName string
	Name        string
}

func (in DNSRecordsClearInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
	)
}

type DNSRecordsClearOutput []DNSRecord
