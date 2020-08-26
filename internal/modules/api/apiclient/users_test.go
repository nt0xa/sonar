package apiclient_test

import (
	"context"
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersCreate(t *testing.T) {
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

func TestUsersDelete(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.UsersDeleteParams{
		Name: "user1",
	}

	res, err := adminClient.UsersDelete(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)
}
