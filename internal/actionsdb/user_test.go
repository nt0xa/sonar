package actionsdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func TestUserCurrent_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	usr, err := acts.ProfileGet(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, "user1", usr.Name)
}

func TestUserCurrent_Error(t *testing.T) {
	_, err := acts.ProfileGet(t.Context())
	assert.Error(t, err)
	assert.IsType(t, &errors.InternalError{}, err)
}
