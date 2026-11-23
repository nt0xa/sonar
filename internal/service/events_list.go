package service

import (
	"context"

	"github.com/nt0xa/sonar/pkg/valid"
)

type EventsList interface {
	EventsList(context.Context, EventsListInput) (EventsListOutput, error)
}

// EventsListMaxCount is the maximum number of events returned per request.
const EventsListMaxCount = 100

// EventsListInput lists Count events after an index oldest first, or before an
// index (latest by default) newest first.
type EventsListInput struct {
	PayloadName string
	Count       uint
	After       *int64
	Before      *int64
	Offset      uint // Deprecated: use Before.
}

func (in EventsListInput) Validate() valid.Problems {
	return valid.Validate(
		valid.String("payloadName", in.PayloadName, valid.Required),
		valid.Number("count", in.Count, valid.Max[uint](EventsListMaxCount)),
		valid.Number("offset", in.Offset,
			valid.Check[uint](in.Offset == 0 || (in.After == nil && in.Before == nil), "cannot be used with after or before")),
		valid.OptionalNumber("after", in.After, valid.Min[int64](0)),
		valid.OptionalNumber("before", in.Before, valid.Min[int64](1),
			valid.Check[int64](in.After == nil, "cannot be used with after")),
	)
}

type EventsListOutput []Event
