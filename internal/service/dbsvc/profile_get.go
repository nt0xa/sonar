package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/service"
)

// ProfileGet implements [service.Service].
func (s *Service) ProfileGet(
	ctx context.Context,
) (*service.ProfileGetOutput, error) {
	id, ok := service.GetUserID(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	u, err := s.db.UsersGetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user(*u), nil
}
