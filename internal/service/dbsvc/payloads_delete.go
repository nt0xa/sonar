package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsDelete implements [service.Service].
func (s *Service) PayloadsDelete(
	ctx context.Context,
	in service.PayloadsDeleteInput,
) (*service.PayloadsDeleteOutput, error) {
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

	p, err = s.db.PayloadsDelete(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	return payload(*p), nil
}
