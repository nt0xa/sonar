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
	return v.Struct(
		v.String("payloadName", in.PayloadName).Required(),
		v.String("method", in.Method).Required().In(HTTPMethodValues()...),
		v.String("path", in.Path).Required().Match(httpPathRegexp, `path must start with "/"`),
		v.Number("code", in.Code).Required(),
	)
}

type HTTPRoutesCreateOutput = HTTPRoute
