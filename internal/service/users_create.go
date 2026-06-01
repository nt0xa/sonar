package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// UsersCreate implements [types.Service].
func (s *service) UsersCreate(
	ctx context.Context,
	in types.UsersCreateInput,
) (*types.UsersCreateOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	_, err := s.db.UsersGetByName(ctx, in.Name)
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("%w: user with name %q already exists", types.ErrConflict, in.Name)
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
