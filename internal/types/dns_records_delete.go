package types

import "context"

type DNSRecordsDelete interface {
	DNSRecordsDelete(context.Context, DNSRecordsDeleteInput) (*DNSRecordsDeleteOutput, error)
}

type DNSRecordsDeleteInput struct {
	PayloadName string
	Index       int64
}

type DNSRecordsDeleteOutput = DNSRecord
