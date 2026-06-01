package types

import "context"

type PayloadsUpdate interface {
	PayloadsUpdate(context.Context, PayloadsUpdateInput) (*PayloadsUpdateOutput, error)
}

type PayloadsUpdateInput struct {
	Name            string
	NewName         string
	NotifyProtocols []ProtoCategory
	StoreEvents     *bool
}

type PayloadsUpdateOutput = Payload
