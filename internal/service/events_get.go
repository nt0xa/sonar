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
	return v.Struct(&in,
		v.String(&in.PayloadName, v.Required),
		v.Int(&in.Index, v.Required),
	)
}

type EventsGetOutput = Event
