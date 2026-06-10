package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/service"
)

// ProfileGet implements [service.Service].
func (s *svc) ProfileGet(
	ctx context.Context,
) (*service.ProfileGetOutput, error) {
	u := getUser(ctx)
	if u == nil {
		return nil, service.Unauthorized()
	}

	return user(*u), nil
}
