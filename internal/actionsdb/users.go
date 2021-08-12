package actionsdb

import (
	"context"
	"database/sql"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func User(m *models.User) *actions.User {
	if m == nil {
		return nil
	}

	return &actions.User{
		Name:      m.Name,
		Params:    m.Params,
		IsAdmin:   m.IsAdmin,
		CreatedAt: m.CreatedAt,
	}
}

func (act *dbactions) UsersCreate(ctx context.Context, p actions.UsersCreateParams) (actions.UsersCreateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	if _, err := act.db.UsersGetByName(p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("user with name %q already exist", p.Name)
	}

	// TODO: check telegram.id and api.token duplicate

	user := &models.User{
		Name:      p.Name,
		Params:    p.Params,
		IsAdmin:   p.IsAdmin,
		CreatedBy: &u.ID,
	}

	if err := act.db.UsersCreate(user); err != nil {
		return nil, errors.Internal(err)
	}

	return User(user), nil
}

func (act *dbactions) UsersDelete(ctx context.Context, p actions.UsersDeleteParams) (actions.UsersDeleteResult, errors.Error) {
	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	user, err := act.db.UsersGetByName(p.Name)
	if err != nil {
		return nil, errors.NotFoundf("user with name %q not found", p.Name)
	}

	if err := act.db.UsersDelete(user.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return User(user), nil
}
