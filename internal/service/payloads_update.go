package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type PayloadsUpdate interface {
	PayloadsUpdate(context.Context, PayloadsUpdateInput) (*PayloadsUpdateOutput, error)
}

type PayloadsUpdateInput struct {
	Name    string
	NewName string
	// Partial update: a nil NotifyProtocols (slice) or StoreEvents (bool needs a
	// pointer since false is meaningful) leaves the existing setting unchanged.
	NotifyProtocols []ProtoCategory
	StoreEvents     *bool
}

func (in PayloadsUpdateInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("name", in.Name, valid.Required),
		valid.Slice("notifyProtocols", in.NotifyProtocols, valid.Each(valid.In(ProtoCategoryValues()...))),
	)
}

type PayloadsUpdateOutput Payload
