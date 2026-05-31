package service

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// authContext looks up a user with get and, if found, returns a context with
// that user attached. Any lookup error is reported as [types.ErrUnauthorized].
func (s *service) authContext(
	ctx context.Context,
	get func(context.Context) (*database.User, error),
) (context.Context, error) {
	u, err := get(ctx)
	if err != nil {
		return nil, types.ErrUnauthorized
	}

	return setUser(ctx, *u), nil
}

// AuthContextByAPIToken implements [types.Service].
func (s *service) AuthContextByAPIToken(ctx context.Context, token string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByAPIToken(ctx, token)
	})
}

// AuthContextByLarkID implements [types.Service].
func (s *service) AuthContextByLarkID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByLarkID(ctx, id)
	})
}

// AuthContextBySlackID implements [types.Service].
func (s *service) AuthContextBySlackID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetBySlackID(ctx, id)
	})
}

// AuthContextByTelegramID implements [types.Service].
func (s *service) AuthContextByTelegramID(ctx context.Context, id int64) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByTelegramID(ctx, id)
	})
}

func (s *service) user(ctx context.Context) *database.User {
	return getUser(ctx)
}
