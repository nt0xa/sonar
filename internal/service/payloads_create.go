package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type PayloadsCreate interface {
	PayloadsCreate(context.Context, PayloadsCreateInput) (*PayloadsCreateOutput, error)
}

type PayloadsCreateInput struct {
	Name            string
	NotifyProtocols []ProtoCategory
	StoreEvents     bool
}

func (in PayloadsCreateInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("name", in.Name, valid.Required),
		valid.Slice("notifyProtocols", in.NotifyProtocols, valid.Each(valid.In(ProtoCategoryValues()...))),
	)
}

type PayloadsCreateOutput Payload
