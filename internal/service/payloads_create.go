package service

import "context"

type PayloadsCreate interface {
	PayloadsCreate(context.Context, PayloadsCreateInput) (*PayloadsCreateOutput, error)
}

type PayloadsCreateInput struct {
	Name            string
	NotifyProtocols []ProtoCategory
	StoreEvents     bool
}

type PayloadsCreateOutput = Payload
