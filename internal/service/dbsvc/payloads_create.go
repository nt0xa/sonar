package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/internal/utils"
)

// PayloadsCreate implements [service.Service].
func (s *svc) PayloadsCreate(
	ctx context.Context,
	in service.PayloadsCreateInput,
) (*service.PayloadsCreateOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	u := s.user(ctx)
	if u == nil {
		return nil, service.Unauthorized()
	}

	_, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.Name)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, service.Conflictf("payload with name %q already exists", in.Name)
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
