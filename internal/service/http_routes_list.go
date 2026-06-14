package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesList interface {
	HTTPRoutesList(context.Context, HTTPRoutesListInput) (HTTPRoutesListOutput, error)
}

type HTTPRoutesListInput struct {
	PayloadName string
}

func (in HTTPRoutesListInput) Validate() v.Problems {
	return v.Validate(
		v.String("payloadName", in.PayloadName, v.Required),
	)
}

type HTTPRoutesListOutput []HTTPRoute
