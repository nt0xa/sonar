package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type HTTPRoutesCreate interface {
	HTTPRoutesCreate(context.Context, HTTPRoutesCreateInput) (*HTTPRoutesCreateOutput, error)
}

type HTTPRoutesCreateInput struct {
	PayloadName string
	Method      HTTPMethod
	Path        string
	Code        int
	Headers     map[string][]string
	Body        string
	IsDynamic   bool
}

func (in HTTPRoutesCreateInput) Validate() v.Problems {
	return v.Validate(
		v.String("payloadName", in.PayloadName, v.Required),
		v.String("method", in.Method, v.Required, v.In(HTTPMethodValues()...)),
		v.String("path", in.Path, v.Required, v.Match(httpPathRegexp, `path must start with "/"`)),
		v.Number("code", in.Code, v.Required),
	)
}

type HTTPRoutesCreateOutput HTTPRoute
