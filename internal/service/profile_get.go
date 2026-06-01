package service

import (
	"context"

	"github.com/nt0xa/sonar/internal/types"
)

// ProfileGet implements [types.Service].
func (s *service) ProfileGet(
	ctx context.Context,
	_ types.ProfileGetInput,
) (*types.ProfileGetOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	return user(*u), nil
}
