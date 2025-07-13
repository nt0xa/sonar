package actionsdb

import (
	"context"
	"database/sql"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func User(m models.User) actions.User {
	return actions.User{
		Name:      m.Name,
		Params:    m.Params,
		IsAdmin:   m.IsAdmin,
		CreatedAt: m.CreatedAt,
	}
}

func (act *dbactions) UsersCreate(ctx context.Context, p actions.UsersCreateParams) (*actions.UsersCreateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	if _, err := act.db.UsersGetByName(ctx, p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("user with name %q already exist", p.Name)
	}

	// TODO: check telegram.id and api.token duplicate

	rec := &models.User{
		Name:      p.Name,
		Params:    p.Params,
		IsAdmin:   p.IsAdmin,
		CreatedBy: &u.ID,
	}

	if err := act.db.UsersCreate(ctx, rec); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.UsersCreateResult{User: User(*rec)}, nil
}

func (act *dbactions) UsersDelete(ctx context.Context, p actions.UsersDeleteParams) (*actions.UsersDeleteResult, errors.Error) {
	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	rec, err := act.db.UsersGetByName(ctx, p.Name)
	if err != nil {
		return nil, errors.NotFoundf("user with name %q not found", p.Name)
	}

	if err := act.db.UsersDelete(ctx, rec.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.UsersDeleteResult{User: User(*rec)}, nil
}
