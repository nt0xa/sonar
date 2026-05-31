package types

import "context"

type HTTPRoutesClear interface {
	HTTPRoutesClear(context.Context, HTTPRoutesClearInput) (HTTPRoutesClearOutput, error)
}

type HTTPRoutesClearInput struct {
	PayloadName string
	Path        string
}

type HTTPRoutesClearOutput = []HTTPRoute
