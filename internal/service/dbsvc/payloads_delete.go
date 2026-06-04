package dbsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// PayloadsDelete implements [service.Service].
func (s *svc) PayloadsDelete(
	ctx context.Context,
	in service.PayloadsDeleteInput,
) (*service.PayloadsDeleteOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, service.ErrUnauthorized
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.Name)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: payload with name %q not found", service.ErrNotFound, in.Name)
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
