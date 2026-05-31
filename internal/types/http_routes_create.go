package types

import "context"

type HTTPRoutesCreate interface {
	HTTPRoutesCreate(context.Context, HTTPRoutesCreateInput) (*HTTPRoutesCreateOutput, error)
}

type HTTPRoutesCreateInput struct {
	PayloadName string
	Method      string
	Path        string
	Code        int
	Headers     map[string][]string
	Body        string
	IsDynamic   bool
}

type HTTPRoutesCreateOutput = HTTPRoute
