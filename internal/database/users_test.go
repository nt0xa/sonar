package database_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database"
)

func TestUsersCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersCreate(t.Context(), database.UsersCreateParams{
		Name:       "test",
		TelegramID: ptr[int64](1234),
	})
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)

	o2, err := db.UsersGetByID(t.Context(), o.ID)
	require.NoError(t, err)
	assert.Equal(t, o, o2)
}

func TestUsersCreate_DuplicateName(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.UsersCreate(t.Context(), database.UsersCreateParams{
		Name: "user1",
	})
	assert.Error(t, err)
}

func TestUsersCreate_DuplicateParam(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.UsersCreate(t.Context(), database.UsersCreateParams{
		Name:       "test",
		TelegramID: ptr[int64](1337),
	})
	assert.Error(t, err)
}

func TestUsersGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)
	assert.Equal(t, "user1", o.Name)
}

func TestUsersGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.UsersGetByID(t.Context(), 1337)
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestUsersGetByName_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByName(t.Context(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, "user1", o.Name)
}

func TestUsersGetByName_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.UsersGetByName(t.Context(), "not-exist")
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestUsersDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.UsersDelete(t.Context(), 1)
	assert.NoError(t, err)
}

func TestUsersUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)
	assert.NotNil(t, o)

	o.TelegramID = ptr[int64](1234)

	updated, err := db.UsersUpdate(t.Context(), database.UsersUpdateParams{
		ID:         o.ID,
		Name:       "user1_updated",
		IsAdmin:    o.IsAdmin,
		TelegramID: ptr[int64](1234),
		CreatedBy:  o.CreatedBy,
	})
	require.NoError(t, err)

	o2, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)
	assert.Equal(t, "user1_updated", o2.Name)
	assert.Equal(t, updated, o2)
}

func TestUsersGetByParams_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByTelegramID(t.Context(), 31337)
	assert.NoError(t, err)
	assert.Equal(t, "user1", o.Name)
}

func TestUsersGetByParams_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.UsersGetByTelegramID(t.Context(), 1)
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}
