package types

import "context"

type DNSRecordsClear interface {
	DNSRecordsClear(context.Context, DNSRecordsClearInput) (DNSRecordsClearOutput, error)
}

type DNSRecordsClearInput struct {
	PayloadName string
	Name        string
}

type DNSRecordsClearOutput = []DNSRecord
