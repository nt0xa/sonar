package dbactions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (act *dbactions) CreateUser(ctx context.Context, p actions.CreateUserParams) (actions.CreateUserResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	if _, err := act.db.UsersGetByName(p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("user with name %q already exist", p.Name)
	}

	user := &models.User{
		Name:      p.Name,
		Params:    p.Params,
		IsAdmin:   p.IsAdmin,
		CreatedBy: &u.ID,
	}

	if err := act.db.UsersCreate(user); err != nil {
		return nil, errors.Internal(err)
	}

	return actions.CreateUserResult(user), nil
}

func (act *dbactions) DeleteUser(ctx context.Context, p actions.DeleteUserParams) (actions.DeleteUserResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

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

	return &actions.MessageResult{Message: fmt.Sprintf("user %q deleted", user.Name)}, nil
}
