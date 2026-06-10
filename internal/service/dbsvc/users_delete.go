package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// UsersDelete implements [service.Service].
func (s *svc) UsersDelete(
	ctx context.Context,
	in service.UsersDeleteInput,
) (*service.UsersDeleteOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	if getUser(ctx) == nil {
		return nil, service.Unauthorized()
	}

	rec, err := s.db.UsersGetByName(ctx, in.Name)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("user with name %q not found", in.Name)
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.UsersDelete(ctx, rec.ID); err != nil {
		return nil, err
	}

	return user(*rec), nil
}
