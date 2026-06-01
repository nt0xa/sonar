package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// PayloadsUpdate implements [types.Service].
func (s *service) PayloadsUpdate(ctx context.Context, in types.PayloadsUpdateInput) (*types.PayloadsUpdateOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.Name)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: payload with name %q not found", types.ErrNotFound, in.Name)
	}
	if err != nil {
		return nil, err
	}

	notifyProtocols := make([]string, len(in.NotifyProtocols))
	for i, np := range in.NotifyProtocols {
		notifyProtocols[i] = string(np)
	}

	params := database.PayloadsUpdateParams{
		ID:              p.ID,
		UserID:          p.UserID,
		Subdomain:       p.Subdomain,
		Name:            p.Name,
		NotifyProtocols: notifyProtocols,
		StoreEvents:     p.StoreEvents,
	}

	if in.NewName != "" {
		params.Name = in.NewName
	}

	if in.StoreEvents != nil {
		params.StoreEvents = *in.StoreEvents
	}

	p, err = s.db.PayloadsUpdate(ctx, params)
	if err != nil {
		return nil, err
	}

	return payload(*p), nil
}
