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
	return v.Struct(
		v.String("payloadName", in.PayloadName).Required(),
		v.String("name", in.Name).Required().Custom(subdomain),
		v.String("type", in.Type).Required().In(DNSRecordTypeValues()...),
		v.StringSlice("values", in.Values).Required().Each().Custom(dnsValueRule(in.Type)),
		v.String("strategy", in.Strategy).Required().In(DNSRecordStrategyValues()...),
	)
}

type DNSRecordsCreateOutput = DNSRecord
