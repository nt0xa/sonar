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
	return v.Struct(
		v.String("payloadName", in.PayloadName).Required(),
	)
}

type HTTPRoutesClearOutput = []HTTPRoute
