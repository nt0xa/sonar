package types

import "context"

type EventsList interface {
	EventsList(context.Context, EventsListInput) (EventsListOutput, error)
}

type EventsListInput struct {
	PayloadName string
	Limit       uint
	Offset      uint
}

type EventsListOutput []Event
