package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// EventsGet implements [types.Service].
func (s *service) EventsGet(
	ctx context.Context,
	in types.EventsGetInput,
) (*types.EventsGetOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: payload with name %q not found", types.ErrNotFound, in.PayloadName)
	}
	if err != nil {
		return nil, err
	}

	r, err := s.db.EventsGetByPayloadAndIndex(ctx, p.ID, in.Index)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: event for payload %q with index %d not found",
			types.ErrNotFound, in.PayloadName, in.Index)
	}
	if err != nil {
		return nil, err
	}

	return event(r.Event, r.Index), nil
}
