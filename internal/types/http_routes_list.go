package types

import "context"

type HTTPRoutesList interface {
	HTTPRoutesList(context.Context, HTTPRoutesListInput) (HTTPRoutesListOutput, error)
}

type HTTPRoutesListInput struct {
	PayloadName string
}

type HTTPRoutesListOutput = []HTTPRoute
