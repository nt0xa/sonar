package service

import "context"

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

type DNSRecordsCreateOutput = DNSRecord
