package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// UsersCreate implements [service.Service].
func (s *Service) UsersCreate(
	ctx context.Context,
	in service.UsersCreateInput,
) (*service.UsersCreateOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	id, ok := service.GetUserID(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	_, err := s.db.UsersGetByName(ctx, in.Name)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, service.Conflictf("user with name %q already exists", in.Name)
	}

	created, err := s.db.UsersCreate(ctx, database.UsersCreateParams{
		Name:       in.Name,
		IsAdmin:    in.IsAdmin,
		CreatedBy:  &id,
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
