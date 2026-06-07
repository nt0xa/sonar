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
	return v.Struct(&in,
		v.String(&in.PayloadName, v.Required),
		v.String(&in.Name, v.Required, v.By(subdomain)),
		v.String(&in.Type, v.Required, v.In(DNSRecordTypeValues()...)),
		v.StringSlice(&in.Values, v.Required, v.Each(dnsValueRule(in.Type))),
		v.String(&in.Strategy, v.Required, v.In(DNSRecordStrategyValues()...)),
	)
}

type DNSRecordsCreateOutput = DNSRecord
