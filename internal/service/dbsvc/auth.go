package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// authContext looks up a user with get and, if found, returns a context with
// that user's id attached. Any lookup error is reported as [service.Unauthorized].
func (s *Service) authContext(
	ctx context.Context,
	get func(context.Context) (*database.User, error),
) (context.Context, error) {
	u, err := get(ctx)
	if err != nil {
		return nil, service.Unauthorized()
	}

	ctx = service.SetUserID(ctx, u.ID)
	ctx = service.SetUserIsAdmin(ctx, u.IsAdmin)

	return ctx, nil
}

// AuthContextByAPIToken implements [service.Service].
func (s *Service) AuthContextByAPIToken(ctx context.Context, token string) (context.Context, error) {
	ctx, err := s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByAPIToken(ctx, token)
	})
	if err != nil {
		return nil, err
	}

	return service.SetSource(ctx, service.AuditSourceApi), nil
}

// AuthContextByLarkID implements [service.Service].
func (s *Service) AuthContextByLarkID(ctx context.Context, id string) (context.Context, error) {
	ctx, err := s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByLarkID(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return service.SetSource(ctx, service.AuditSourceLark), nil
}

// AuthContextBySlackID implements [service.Service].
func (s *Service) AuthContextBySlackID(ctx context.Context, id string) (context.Context, error) {
	ctx, err := s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetBySlackID(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return service.SetSource(ctx, service.AuditSourceSlack), nil
}

// AuthContextByTelegramID implements [service.Service].
func (s *Service) AuthContextByTelegramID(ctx context.Context, id int64) (context.Context, error) {
	ctx, err := s.authContext(ctx, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByTelegramID(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return service.SetSource(ctx, service.AuditSourceTelegram), nil
}
