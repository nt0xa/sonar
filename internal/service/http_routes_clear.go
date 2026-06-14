package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesClear interface {
	HTTPRoutesClear(context.Context, HTTPRoutesClearInput) (HTTPRoutesClearOutput, error)
}

type HTTPRoutesClearInput struct {
	PayloadName string
	Path        string
}

func (in HTTPRoutesClearInput) Validate() v.Problems {
	return v.Validate(
		v.String("payloadName", in.PayloadName, v.Required),
	)
}

type HTTPRoutesClearOutput []HTTPRoute
