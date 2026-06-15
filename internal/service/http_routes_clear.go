package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesClear interface {
	HTTPRoutesClear(context.Context, HTTPRoutesClearInput) (HTTPRoutesClearOutput, error)
}

type HTTPRoutesClearInput struct {
	PayloadName string
	Path        string
}

func (in HTTPRoutesClearInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
	)
}

type HTTPRoutesClearOutput []HTTPRoute
