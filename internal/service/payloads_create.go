package service

import (
	"context"
	"errors"
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

	_, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.Name)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("%w: payload with name %q already exists", types.ErrConflict, in.Name)
	}

	subdomain, err := utils.GenerateRandomString(4)
	if err != nil {
		return nil, err
	}

	notifyProtocols := make([]string, len(in.NotifyProtocols))
	for i, p := range in.NotifyProtocols {
		notifyProtocols[i] = string(p)
	}

	p, err := s.db.PayloadsCreate(ctx, database.PayloadsCreateParams{
		UserID:          u.ID,
		Subdomain:       subdomain,
		Name:            in.Name,
		NotifyProtocols: notifyProtocols,
		StoreEvents:     in.StoreEvents,
	})
	if err != nil {
		return nil, err
	}

	return payload(*p), nil
}
