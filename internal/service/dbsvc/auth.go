package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// authContext looks up a user with get and, if found, returns a context with
// that user attached. Any lookup error is reported as [service.Unauthorized].
func (s *Service) authContext(
	ctx context.Context,
	get func(context.Context) (*database.User, error),
) (context.Context, error) {
	u, err := get(ctx)
	if err != nil {
		return nil, service.Unauthorized()
	}

	return setUser(ctx, *u), nil
}

// AuthContextByAPIToken implements [service.Service].
func (s *Service) AuthContextByAPIToken(ctx context.Context, token string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByAPIToken(ctx, token)
	})
}

// AuthContextByLarkID implements [service.Service].
func (s *Service) AuthContextByLarkID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByLarkID(ctx, id)
	})
}

// AuthContextBySlackID implements [service.Service].
func (s *Service) AuthContextBySlackID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetBySlackID(ctx, id)
	})
}

// AuthContextByTelegramID implements [service.Service].
func (s *Service) AuthContextByTelegramID(ctx context.Context, id int64) (context.Context, error) {
	return s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByTelegramID(ctx, id)
	})
}
