package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type PayloadsDelete interface {
	PayloadsDelete(context.Context, PayloadsDeleteInput) (*PayloadsDeleteOutput, error)
}

type PayloadsDeleteInput struct {
	Name string
}

func (in PayloadsDeleteInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("name", in.Name, valid.Required),
	)
}

type PayloadsDeleteOutput Payload
