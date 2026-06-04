package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/service"
)

// ProfileGet implements [service.Service].
func (s *svc) ProfileGet(
	ctx context.Context,
	_ service.ProfileGetInput,
) (*service.ProfileGetOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, service.ErrUnauthorized
	}

	return user(*u), nil
}
