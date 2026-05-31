package types

import "context"

type DNSRecordsList interface {
	DNSRecordsList(context.Context, DNSRecordsListInput) (DNSRecordsListOutput, error)
}

type DNSRecordsListInput struct {
	PayloadName string
}

type DNSRecordsListOutput = []DNSRecord
