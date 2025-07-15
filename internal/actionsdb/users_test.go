package actionsdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func TestUsersCreate_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		p    actions.UsersCreateParams
	}{
		{
			"regular",
			actions.UsersCreateParams{
				Name: "test",
				Params: models.UserParams{
					TelegramID: 1000,
					APIToken:   "token",
				},
			},
		},
		{
			"admin",
			actions.UsersCreateParams{
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

			r, err := acts.UsersCreate(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
			assert.Equal(t, tt.p.Name, r.Name)
			assert.Equal(t, int64(tt.p.Params.TelegramID), r.Params.TelegramID)
			assert.Equal(t, tt.p.Params.APIToken, r.Params.APIToken)
			assert.Equal(t, tt.p.IsAdmin, r.IsAdmin)
		})
	}
}

func TestUsersCreate_Error(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.UsersCreateParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			t.Context(),
			actions.UsersCreateParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty name",
			ctx,
			actions.UsersCreateParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"duplicate name",
			ctx,
			actions.UsersCreateParams{
				Name: "user1",
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.UsersCreate(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDeleteUser_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 3)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		p    actions.UsersDeleteParams
	}{
		{
			"user1",
			actions.UsersDeleteParams{
				Name: "user1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.UsersDelete(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
		})
	}
}

func TestDeleteUser_Error(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.UsersDeleteParams
		err  errors.Error
	}{
		{
			"empty name",
			ctx,
			actions.UsersDeleteParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing user",
			ctx,
			actions.UsersDeleteParams{
				Name: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.UsersDelete(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
