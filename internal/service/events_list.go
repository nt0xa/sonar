package service

import (
	"context"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type EventsList interface {
	EventsList(context.Context, EventsListInput) (EventsListOutput, error)
}

type EventsListInput struct {
	PayloadName string
	Limit       uint
	Offset      uint
}

func (in EventsListInput) Validate() v.Problems {
	return v.Struct(&in,
		v.String(&in.PayloadName, v.Required),
	)
}

type EventsListOutput = []Event
