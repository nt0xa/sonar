package dbsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// EventsGet implements [service.Service].
func (s *svc) EventsGet(
	ctx context.Context,
	in service.EventsGetInput,
) (*service.EventsGetOutput, error) {
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

	r, err := s.db.EventsGetByPayloadAndIndex(ctx, p.ID, in.Index)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: event for payload %q with index %d not found",
			service.ErrNotFound, in.PayloadName, in.Index)
	}
	if err != nil {
		return nil, err
	}

	return event(r.Event, r.Index), nil
}
