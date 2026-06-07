package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type PayloadsDelete interface {
	PayloadsDelete(context.Context, PayloadsDeleteInput) (*PayloadsDeleteOutput, error)
}

type PayloadsDeleteInput struct {
	Name string
}

func (in PayloadsDeleteInput) Validate() v.Problems {
	return v.Struct(
		v.String("name", in.Name).Required(),
	)
}

type PayloadsDeleteOutput = Payload
