package service

import (
	"context"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
	"github.com/nt0xa/sonar/internal/utils"
)

// PayloadsCreate implements [types.Service].
func (s *service) PayloadsCreate(
	ctx context.Context,
	in types.PayloadsCreateInput,
) (*types.PayloadsCreateOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	if _, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.Name); err != database.ErrNoRows {
		return nil, fmt.Errorf("%w: payload with name %q already exist", types.ErrConflict, in.Name)
	}

	subdomain, err := utils.GenerateRandomString(4)
	if err != nil {
		return nil, err
	}

	p, err := s.db.PayloadsCreate(ctx, database.PayloadsCreateParams{
		UserID:          u.ID,
		Subdomain:       subdomain,
		Name:            in.Name,
		NotifyProtocols: in.NotifyProtocols,
		StoreEvents:     in.StoreEvents,
	})
	if err != nil {
		return nil, err
	}

	return payload(*p), nil
}
