package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type PayloadsUpdate interface {
	PayloadsUpdate(context.Context, PayloadsUpdateInput) (*PayloadsUpdateOutput, error)
}

type PayloadsUpdateInput struct {
	Name            string
	NewName         string
	NotifyProtocols []ProtoCategory
	StoreEvents     *bool
}

func (in PayloadsUpdateInput) Validate() v.Problems {
	return v.Struct(&in,
		v.String(&in.Name, v.Required),
		v.StringSlice(&in.NotifyProtocols, v.Each(v.In(ProtoCategoryValues()...))),
	)
}

type PayloadsUpdateOutput = Payload
