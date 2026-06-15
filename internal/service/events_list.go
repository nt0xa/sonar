package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type EventsList interface {
	EventsList(context.Context, EventsListInput) (EventsListOutput, error)
}

type EventsListInput struct {
	PayloadName string
	Limit       uint
	Offset      uint
}

func (in EventsListInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
	)
}

type EventsListOutput []Event
