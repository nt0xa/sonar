package dbsvc

import (
	"cmp"
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// EventsList implements [service.Service].
func (s *svc) EventsList(
	ctx context.Context,
	in service.EventsListInput,
) (service.EventsListOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	u := s.user(ctx)
	if u == nil {
		return nil, service.Unauthorized()
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("payload with name %q not found", in.PayloadName)
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
