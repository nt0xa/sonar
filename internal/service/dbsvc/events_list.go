package dbsvc

import (
	"cmp"
	"context"
	"errors"
	"math"
	"slices"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// EventsList implements [service.Service].
func (s *Service) EventsList(
	ctx context.Context,
	in service.EventsListInput,
) (service.EventsListOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	c, ok := service.CallerFrom(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, c.UserID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("payload with name %q not found", in.PayloadName)
	}
	if err != nil {
		return nil, err
	}

	count := int64(cmp.Or(in.Count, 10))

	// Events after a cursor, or the first ones with Reverse, are selected oldest
	// first; the rest newest first. Results are then put in display order.
	asc := in.After != nil || (in.Reverse && in.Before == nil)

	var recs []*database.Event
	if asc {
		var after int64
		if in.After != nil {
			after = *in.After
		}
		recs, err = s.db.EventsListByPayloadIDAfter(ctx, database.EventsListByPayloadIDAfterParams{
			PayloadID: p.ID,
			After:     after,
			PageLimit: count,
		})
	} else {
		before := int64(math.MaxInt64)
		if in.Before != nil {
			before = *in.Before
		}
		recs, err = s.db.EventsListByPayloadID(ctx, database.EventsListByPayloadIDParams{
			PayloadID:  p.ID,
			Before:     before,
			PageLimit:  count,
			PageOffset: int64(in.Offset),
		})
	}
	if err != nil {
		return nil, err
	}

	if asc != in.Reverse {
		slices.Reverse(recs)
	}

	out := make([]service.Event, len(recs))

	for i, r := range recs {
		out[i] = *event(*r)
	}

	return out, nil
}
