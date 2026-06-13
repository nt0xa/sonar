package dbsvc

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// authContext looks up a user with get and, if found, returns a context
// carrying that user's [service.Caller] identity. Any lookup error is reported
// as [service.Unauthorized].
func (s *Service) authContext(
	ctx context.Context,
	source service.AuditSource,
	get func(context.Context) (*database.User, error),
) (context.Context, error) {
	u, err := get(ctx)
	if err != nil {
		return nil, service.Unauthorized()
	}

	return service.WithCaller(ctx, service.Caller{
		UserID:     u.ID,
		UserName:   u.Name,
		IsAdmin:    u.IsAdmin,
		Source:     source,
		TelegramID: u.TelegramID,
		LarkID:     u.LarkID,
		SlackID:    u.SlackID,
	}), nil
}

// AuthContextByAPIToken implements [service.Service].
func (s *Service) AuthContextByAPIToken(ctx context.Context, token string) (context.Context, error) {
	return s.authContext(ctx, service.AuditSourceApi, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByAPIToken(ctx, token)
	})
}

// AuthContextByLarkID implements [service.Service].
func (s *Service) AuthContextByLarkID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, service.AuditSourceLark, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByLarkID(ctx, id)
	})
}

// AuthContextBySlackID implements [service.Service].
func (s *Service) AuthContextBySlackID(ctx context.Context, id string) (context.Context, error) {
	return s.authContext(ctx, service.AuditSourceSlack, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetBySlackID(ctx, id)
	})
}

// AuthContextByTelegramID implements [service.Service].
func (s *Service) AuthContextByTelegramID(ctx context.Context, id int64) (context.Context, error) {
	return s.authContext(ctx, service.AuditSourceTelegram, func(ctx context.Context) (*database.User, error) {
		return s.db.UsersGetByTelegramID(ctx, id)
	})
}
