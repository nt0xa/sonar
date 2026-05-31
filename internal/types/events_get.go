package types

import "context"

type EventsGet interface {
	EventsGet(context.Context, EventsGetInput) (*EventsGetOutput, error)
}

type EventsGetInput struct {
	PayloadName string
	Index       int64
}

type EventsGetOutput Event
