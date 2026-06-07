package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
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

func (in HTTPRoutesUpdateInput) Validate() v.Problems {
	return v.Validate(
		v.String("payload", in.Payload, v.Required),
		v.OptionalString("method", in.Method, v.In(HTTPMethodValues()...)),
		v.OptionalString("path", in.Path, v.Match(httpPathRegexp, `path must start with "/"`)),
	)
}

type HTTPRoutesUpdateOutput = HTTPRoute
