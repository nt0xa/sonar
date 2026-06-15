package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type DNSRecordsCreate interface {
	DNSRecordsCreate(context.Context, DNSRecordsCreateInput) (*DNSRecordsCreateOutput, error)
}

type DNSRecordsCreateInput struct {
	PayloadName string
	Name        string
	TTL         int
	Type        DNSRecordType
	Values      []string
	Strategy    DNSRecordStrategy
}

func (in DNSRecordsCreateInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
		valid.String("name", in.Name, valid.Required, subdomain),
		valid.String("type", in.Type, valid.Required, valid.In(DNSRecordTypeValues()...)),
		valid.Slice("values", in.Values, valid.NotEmpty, valid.Each(dnsValueRule(in.Type))),
		valid.String("strategy", in.Strategy, valid.Required, valid.In(DNSRecordStrategyValues()...)),
	)
}

type DNSRecordsCreateOutput DNSRecord
