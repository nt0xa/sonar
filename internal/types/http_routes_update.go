package types

import "context"

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

type HTTPRoutesUpdateOutput = HTTPRoute
