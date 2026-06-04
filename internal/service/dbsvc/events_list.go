package dbsvc

import (
	"cmp"
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// EventsList implements [service.Service].
func (s *svc) EventsList(
	ctx context.Context,
	in service.EventsListInput,
) (service.EventsListOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, service.ErrUnauthorized
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: payload with name %q not found", service.ErrNotFound, in.PayloadName)
	}
	if err != nil {
		return nil, err
	}

	limit := cmp.Or(in.Limit, 10)

	recs, err := s.db.EventsListByPayloadID(ctx, database.EventsListByPayloadIDParams{
		PayloadID: p.ID,
		Limit:     int64(limit),
		Offset:    int64(in.Offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]service.Event, len(recs))

	for i, r := range recs {
		out[i] = *event(r.Event, r.Index)
	}

	return out, nil
}
