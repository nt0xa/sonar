package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
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

func (in DNSRecordsCreateInput) Validate() v.Problems {
	return v.Validate(
		v.String("payloadName", in.PayloadName, v.Required),
		v.String("name", in.Name, v.Required, subdomain),
		v.String("type", in.Type, v.Required, v.In(DNSRecordTypeValues()...)),
		v.Slice("values", in.Values, v.NotEmpty, v.Each(dnsValueRule(in.Type))),
		v.String("strategy", in.Strategy, v.Required, v.In(DNSRecordStrategyValues()...)),
	)
}

type DNSRecordsCreateOutput = DNSRecord
