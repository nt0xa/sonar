package dbsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// UsersCreate implements [service.Service].
func (s *svc) UsersCreate(
	ctx context.Context,
	in service.UsersCreateInput,
) (*service.UsersCreateOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, service.ErrUnauthorized
	}

	_, err := s.db.UsersGetByName(ctx, in.Name)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("%w: user with name %q already exists", service.ErrConflict, in.Name)
	}

	created, err := s.db.UsersCreate(ctx, database.UsersCreateParams{
		Name:       in.Name,
		IsAdmin:    in.IsAdmin,
		CreatedBy:  &u.ID,
		APIToken:   in.APIToken,
		TelegramID: in.TelegramID,
		LarkID:     in.LarkID,
		SlackID:    in.SlackID,
	})
	if err != nil {
		return nil, err
	}

	return user(*created), nil
}
