package service

import (
	"context"
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
	if err == database.ErrNoRows {
		return nil, fmt.Errorf("%w: payload with name %q not found", types.ErrNotFound, in.Name)
	} else if err != nil {
		return nil, err
	}

	updateParams := database.PayloadsUpdateParams{
		ID:              p.ID,
		UserID:          p.UserID,
		Subdomain:       p.Subdomain,
		Name:            p.Name,
		NotifyProtocols: in.NotifyProtocols,
		StoreEvents:     p.StoreEvents,
	}

	if in.NewName != "" {
		updateParams.Name = in.NewName
	}

	if in.StoreEvents != nil {
		updateParams.StoreEvents = *in.StoreEvents
	}

	p, err = s.db.PayloadsUpdate(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return payload(*p), nil
}
