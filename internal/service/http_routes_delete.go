package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesDelete interface {
	HTTPRoutesDelete(context.Context, HTTPRoutesDeleteInput) (*HTTPRoutesDeleteOutput, error)
}

type HTTPRoutesDeleteInput struct {
	PayloadName string
	Index       int64
}

func (in HTTPRoutesDeleteInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
		valid.Number("index", in.Index, valid.Required),
	)
}

type HTTPRoutesDeleteOutput HTTPRoute
