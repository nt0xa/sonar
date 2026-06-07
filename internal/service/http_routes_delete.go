package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesDelete interface {
	HTTPRoutesDelete(context.Context, HTTPRoutesDeleteInput) (*HTTPRoutesDeleteOutput, error)
}

type HTTPRoutesDeleteInput struct {
	PayloadName string
	Index       int64
}

func (in HTTPRoutesDeleteInput) Validate() v.Problems {
	return v.Struct(
		v.String("payloadName", in.PayloadName).Required(),
		v.Number("index", in.Index).Required(),
	)
}

type HTTPRoutesDeleteOutput = HTTPRoute
