package dbsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// UsersDelete implements [service.Service].
func (s *svc) UsersDelete(
	ctx context.Context,
	in service.UsersDeleteInput,
) (*service.UsersDeleteOutput, error) {
	if s.user(ctx) == nil {
		return nil, service.ErrUnauthorized
	}

	rec, err := s.db.UsersGetByName(ctx, in.Name)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: user with name %q not found", service.ErrNotFound, in.Name)
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.UsersDelete(ctx, rec.ID); err != nil {
		return nil, err
	}

	return user(*rec), nil
}
