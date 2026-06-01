package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// UsersDelete implements [types.Service].
func (s *service) UsersDelete(
	ctx context.Context,
	in types.UsersDeleteInput,
) (*types.UsersDeleteOutput, error) {
	if s.user(ctx) == nil {
		return nil, types.ErrUnauthorized
	}

	rec, err := s.db.UsersGetByName(ctx, in.Name)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: user with name %q not found", types.ErrNotFound, in.Name)
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.UsersDelete(ctx, rec.ID); err != nil {
		return nil, err
	}

	return user(*rec), nil
}
