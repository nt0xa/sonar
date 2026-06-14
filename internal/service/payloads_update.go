package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
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

func (in PayloadsUpdateInput) Validate() v.Problems {
	return v.Validate(
		v.String("name", in.Name, v.Required),
		v.Slice("notifyProtocols", in.NotifyProtocols, v.Each(v.In(ProtoCategoryValues()...))),
	)
}

type PayloadsUpdateOutput Payload
