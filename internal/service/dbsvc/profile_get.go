package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/service"
)

// ProfileGet implements [service.Service].
func (s *Service) ProfileGet(
	ctx context.Context,
) (*service.ProfileGetOutput, error) {
	c, ok := service.CallerFrom(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	u, err := s.db.UsersGetByID(ctx, c.UserID)
	if err != nil {
		return nil, err
	}

	return user(*u), nil
}
