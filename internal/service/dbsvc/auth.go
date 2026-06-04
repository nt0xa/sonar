package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// authContext looks up a user with get and, if found, returns a context with
// that user attached. Any lookup error is reported as [service.ErrUnauthorized].
func (s *svc) authContext(
	ctx context.Context,
	get func(context.Context) (*database.User, error),
) (context.Context, error) {
	u, err := get(ctx)
	if err != nil {
		return nil, service.ErrUnauthorized
	}

	return setUser(ctx, *u), nil
}

// AuthContextByAPIToken implements [service.Service].
func (s *svc) AuthContextByAPIToken(ctx context.Context, token string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByAPIToken(ctx, token)
	})
}

// AuthContextByLarkID implements [service.Service].
func (s *svc) AuthContextByLarkID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByLarkID(ctx, id)
	})
}

// AuthContextBySlackID implements [service.Service].
func (s *svc) AuthContextBySlackID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetBySlackID(ctx, id)
	})
}

// AuthContextByTelegramID implements [service.Service].
func (s *svc) AuthContextByTelegramID(ctx context.Context, id int64) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByTelegramID(ctx, id)
	})
}

func (s *svc) user(ctx context.Context) *database.User {
	return getUser(ctx)
}
