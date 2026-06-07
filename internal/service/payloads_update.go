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
	return v.Struct(
		v.String("name", in.Name).Required(),
		v.StringSlice("notifyProtocols", in.NotifyProtocols).Each().In(ProtoCategoryValues()...),
	)
}

type PayloadsUpdateOutput = Payload
