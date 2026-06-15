package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
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

func (in HTTPRoutesCreateInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
		valid.String("method", in.Method, valid.Required, valid.In(HTTPMethodValues()...)),
		valid.String("path", in.Path, valid.Required, valid.Match(httpPathRegexp, `path must start with "/"`)),
		valid.Number("code", in.Code, valid.Required),
	)
}

type HTTPRoutesCreateOutput HTTPRoute
