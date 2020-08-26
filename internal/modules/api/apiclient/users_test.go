package apiclient_test

import (
	"context"
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.UsersCreateParams{
		Name: "test",
	}

	res, err := adminClient.UsersCreate(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, p.Name, res.Name)
}

func TestUsersCreate_Error(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.UsersCreateParams{
		Name: "",
	}

	res, err := adminClient.UsersCreate(context.Background(), p)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestUsersDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.UsersDeleteParams{
		Name: "user1",
	}

	res, err := adminClient.UsersDelete(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestUsersDelete_Error(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.UsersDeleteParams{
		Name: "not-exist",
	}

	res, err := adminClient.UsersDelete(context.Background(), p)
	require.Error(t, err)
	require.Nil(t, res)
}
