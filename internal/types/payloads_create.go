package types

import "context"

type PayloadsCreate interface {
	PayloadsCreate(context.Context, PayloadsCreateInput) (*PayloadsCreateOutput, error)
}

type PayloadsCreateInput struct {
	Name            string
	NotifyProtocols []string
	StoreEvents     bool
}

type PayloadsCreateOutput Payload
