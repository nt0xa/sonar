package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesList interface {
	HTTPRoutesList(context.Context, HTTPRoutesListInput) (HTTPRoutesListOutput, error)
}

type HTTPRoutesListInput struct {
	PayloadName string
}

func (in HTTPRoutesListInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
	)
}

type HTTPRoutesListOutput []HTTPRoute
