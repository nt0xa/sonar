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
	return v.Validate(
		v.String("name", in.Name, v.Required),
	)
}

type PayloadsDeleteOutput = Payload
