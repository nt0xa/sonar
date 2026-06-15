package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type EventsGet interface {
	EventsGet(context.Context, EventsGetInput) (*EventsGetOutput, error)
}

type EventsGetInput struct {
	PayloadName string
	Index       int64
}

func (in EventsGetInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
		valid.Number("index", in.Index, valid.Required),
	)
}

type EventsGetOutput Event
