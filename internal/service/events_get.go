package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type EventsGet interface {
	EventsGet(context.Context, EventsGetInput) (*EventsGetOutput, error)
}

type EventsGetInput struct {
	PayloadName string
	Index       int64
}

func (in EventsGetInput) Validate() v.Problems {
	return v.Struct(
		v.String("payloadName", in.PayloadName).Required(),
		v.Number("index", in.Index).Required(),
	)
}

type EventsGetOutput = Event
