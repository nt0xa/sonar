package types

import "context"

type HTTPRoutesDelete interface {
	HTTPRoutesDelete(context.Context, HTTPRoutesDeleteInput) (*HTTPRoutesDeleteOutput, error)
}

type HTTPRoutesDeleteInput struct {
	PayloadName string
	Index       int64
}

type HTTPRoutesDeleteOutput = HTTPRoute
