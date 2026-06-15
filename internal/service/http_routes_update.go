package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesUpdate interface {
	HTTPRoutesUpdate(context.Context, HTTPRoutesUpdateInput) (*HTTPRoutesUpdateOutput, error)
}

type HTTPRoutesUpdateInput struct {
	Payload   string
	Index     int64
	Method    *HTTPMethod
	Path      *string
	Code      *int
	Headers   map[string][]string
	Body      *string
	IsDynamic *bool
}

func (in HTTPRoutesUpdateInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payload", in.Payload, valid.Required),
		valid.OptionalString("method", in.Method, valid.In(HTTPMethodValues()...)),
		valid.OptionalString("path", in.Path, valid.Match(httpPathRegexp, `path must start with "/"`)),
	)
}

type HTTPRoutesUpdateOutput HTTPRoute
