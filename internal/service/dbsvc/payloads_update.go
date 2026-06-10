package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsUpdate implements [service.Service].
func (s *svc) PayloadsUpdate(ctx context.Context, in service.PayloadsUpdateInput) (*service.PayloadsUpdateOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	u := getUser(ctx)
	if u == nil {
		return nil, service.Unauthorized()
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.Name)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("payload with name %q not found", in.Name)
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
