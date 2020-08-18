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
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.CreateUserParams{
		Name: "test",
		Params: models.UserParams{
			TelegramID: 1000,
		},
		IsAdmin: true,
	}

	r, err := acts.CreateUser(ctx, p)
	assert.NoError(t, err)

	assert.NotNil(t, r)
	assert.Equal(t, "test", r.Name)
	assert.Equal(t, int64(1000), r.Params.TelegramID)
	assert.Equal(t, true, r.IsAdmin)
}

func TestCreateUser_Validation(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.CreateUserParams{
		Name: "",
	}

	_, err = acts.CreateUser(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.ValidationError{}, err)
}

func TestCreateUser_Conflict(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.CreateUserParams{
		Name: "user2",
	}

	_, err = acts.CreateUser(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.ConflictError{}, err)
}

func TestDeleteUser_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.DeleteUserParams{
		Name: "user2",
	}

	_, err = acts.DeleteUser(ctx, p)
	assert.NoError(t, err)
}

func TestDeleteUser_Validation(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.DeleteUserParams{
		Name: "",
	}

	_, err = acts.DeleteUser(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.ValidationError{}, err)
}

func TestDeleteUser_NotFound(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.DeleteUserParams{
		Name: "not-exist",
	}

	_, err = acts.DeleteUser(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.NotFoundError{}, err)
}
