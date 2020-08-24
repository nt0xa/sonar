package dbactions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func TestCreateUser_Success(t *testing.T) {
	u, err := db.UsersGetByID(3) // admin
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		p    actions.CreateUserParams
	}{
		{
			"regular",
			actions.CreateUserParams{
				Name: "test",
				Params: models.UserParams{
					TelegramID: 1000,
					APIToken:   "token",
				},
			},
		},
		{
			"admin",
			actions.CreateUserParams{
				Name: "test",
				Params: models.UserParams{
					TelegramID: 1000,
					APIToken:   "token",
				},
				IsAdmin: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.CreateUser(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
			assert.Equal(t, tt.p.Name, r.Name)
			assert.Equal(t, int64(tt.p.Params.TelegramID), r.Params.TelegramID)
			assert.Equal(t, tt.p.Params.APIToken, r.Params.APIToken)
			assert.Equal(t, tt.p.IsAdmin, r.IsAdmin)
		})
	}
}

func TestCreateUser_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	adm, err := db.UsersGetByID(3)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), adm)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.CreateUserParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.CreateUserParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"not admin",
			actions.SetUser(context.Background(), u),
			actions.CreateUserParams{
				Name: "test",
			},
			&errors.ForbiddenError{},
		},
		{
			"empty name",
			ctx,
			actions.CreateUserParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"duplicate name",
			ctx,
			actions.CreateUserParams{
				Name: "user1",
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.CreateUser(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDeleteUser_Success(t *testing.T) {
	u, err := db.UsersGetByID(3)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		p    actions.DeleteUserParams
	}{
		{
			"user1",
			actions.DeleteUserParams{
				Name: "user1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.DeleteUser(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
		})
	}
}

func TestDeleteUser_Error(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	adm, err := db.UsersGetByID(3)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), adm)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.DeleteUserParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.DeleteUserParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"not admin",
			actions.SetUser(context.Background(), u),
			actions.DeleteUserParams{
				Name: "test",
			},
			&errors.ForbiddenError{},
		},
		{
			"empty name",
			ctx,
			actions.DeleteUserParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing user",
			ctx,
			actions.DeleteUserParams{
				Name: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DeleteUser(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
