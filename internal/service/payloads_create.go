package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type PayloadsCreate interface {
	PayloadsCreate(context.Context, PayloadsCreateInput) (*PayloadsCreateOutput, error)
}

type PayloadsCreateInput struct {
	Name            string
	NotifyProtocols []ProtoCategory
	StoreEvents     bool
}

func (in PayloadsCreateInput) Validate() v.Problems {
	return v.Struct(&in,
		v.String(&in.Name, v.Required),
		v.StringSlice(&in.NotifyProtocols, v.Each(v.In(ProtoCategoryValues()...))),
	)
}

type PayloadsCreateOutput = Payload
