package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// EventsGet implements [service.Service].
func (s *svc) EventsGet(
	ctx context.Context,
	in service.EventsGetInput,
) (*service.EventsGetOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	u := getUser(ctx)
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

	r, err := s.db.EventsGetByPayloadAndIndex(ctx, p.ID, in.Index)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("event for payload %q with index %d not found",
			in.PayloadName, in.Index)
	}
	if err != nil {
		return nil, err
	}

	return event(r.Event, r.Index), nil
}
