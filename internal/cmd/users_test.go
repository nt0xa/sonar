package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
)

func TestCreateUser_Success(t *testing.T) {
	cmd, acts, hnd := prepare()

	res := user

	acts.
		On("CreateUser", admin, actions.CreateUserParams{
			Name: "user",
			Params: models.UserParams{
				TelegramID: 1337,
				APIToken:   "token",
			},
		}).
		Return(res, nil)

	hnd.
		On("Handle", mock.Anything, res)

	_, err := cmd.Exec(context.Background(), admin,
		strings.Split("users new user -p telegram.id=1337 -p api.token=token", " "))

	assert.NoError(t, err)
}

func TestCreateUser_NoArg(t *testing.T) {
	cmd, _, _ := prepare()

	out, err := cmd.Exec(context.Background(), admin, []string{"users", "new"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestDeleteUser_Success(t *testing.T) {
	cmd, acts, hnd := prepare()

	res := &actions.MessageResult{Message: "test"}

	acts.
		On("DeleteUser", admin, actions.DeleteUserParams{
			Name: "user",
		}).
		Return(res, nil)

	hnd.
		On("Handle", mock.Anything, res)

	_, err := cmd.Exec(context.Background(), admin, []string{"users", "del", "user"})
	assert.NoError(t, err)
}

func TestDeleteUser_NoArg(t *testing.T) {
	cmd, _, _ := prepare()

	out, err := cmd.Exec(context.Background(), admin, []string{"users", "del"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}
