package types

import "context"

type DNSRecordsCreate interface {
	DNSRecordsCreate(context.Context, DNSRecordsCreateInput) (*DNSRecordsCreateOutput, error)
}

type DNSRecordsCreateInput struct {
	PayloadName string
	Name        string
	TTL         int
	Type        string
	Values      []string
	Strategy    string
}

type DNSRecordsCreateOutput DNSRecord
