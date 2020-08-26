package apiclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserCurrent(t *testing.T) {
	setup(t)
	defer teardown(t)

	res, err := client.UserCurrent(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res)
}
