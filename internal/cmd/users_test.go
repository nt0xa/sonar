package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
)

func TestCreateUser_Success(t *testing.T) {
	c, acts, hnd := prepare()

	res := actions.UsersCreateResult(&actions.User{})

	acts.
		On("UsersCreate", ctx, actions.UsersCreateParams{
			Name: "user",
			Params: models.UserParams{
				TelegramID: 1337,
				APIToken:   "token",
			},
		}).
		Return(res, nil)

	hnd.
		On("UsersCreate", ctx, res)

	_, err := c.Exec(ctx, &actions.User{IsAdmin: true},
		strings.Split("users new user -p telegram.id=1337 -p api.token=token", " "))

	assert.NoError(t, err)

	acts.AssertExpectations(t)
	hnd.AssertExpectations(t)
}

func TestCreateUser_NoArg(t *testing.T) {
	c, _, _ := prepare()

	out, err := c.Exec(context.Background(), &actions.User{IsAdmin: true}, []string{"users", "new"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestDeleteUser_Success(t *testing.T) {
	c, acts, hnd := prepare()

	res := actions.UsersDeleteResult(&actions.User{})

	acts.
		On("UsersDelete", ctx, actions.UsersDeleteParams{
			Name: "user",
		}).
		Return(res, nil)

	hnd.
		On("UsersDelete", ctx, res)

	_, err := c.Exec(ctx, &actions.User{IsAdmin: true}, []string{"users", "del", "user"})
	assert.NoError(t, err)

	acts.AssertExpectations(t)
	hnd.AssertExpectations(t)
}

func TestDeleteUser_NoArg(t *testing.T) {
	c, _, _ := prepare()

	out, err := c.Exec(context.Background(), &actions.User{IsAdmin: true}, []string{"users", "del"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")

}
