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
	return v.Struct(
		v.String("name", in.Name).Required(),
		v.StringSlice("notifyProtocols", in.NotifyProtocols).Each().In(ProtoCategoryValues()...),
	)
}

type PayloadsCreateOutput = Payload
