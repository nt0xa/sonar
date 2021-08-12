package database_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/database/models"
)

func TestUsersCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.User{
		Name: "test",
		Params: models.UserParams{
			TelegramID: 1234,
		},
	}

	err := db.UsersCreate(o)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)

	o2, err := db.UsersGetByID(o.ID)
	require.NoError(t, err)
	assert.Equal(t, o.Params, o2.Params)
}

func TestUsersCreate_DuplicateName(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.User{
		Name: "user1",
	}

	err := db.UsersCreate(o)
	assert.Error(t, err)
}

func TestUsersCreate_DuplicateParam(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.User{
		Name: "test",
		Params: models.UserParams{
			TelegramID: 1337,
		},
	}

	err := db.UsersCreate(o)
	assert.Error(t, err)
}

func TestUsersGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, o)
	assert.Equal(t, "user1", o.Name)
}

func TestUsersGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByID(1337)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestUsersGetByName_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByName("user1")
	assert.NoError(t, err)
	assert.NotNil(t, o)
	assert.Equal(t, "user1", o.Name)
}

func TestUsersGetByName_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.UsersGetByName("not-exist")
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestUsersDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.UsersDelete(1)
	assert.NoError(t, err)
}

func TestUsersUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByID(1)
	require.NoError(t, err)
	assert.NotNil(t, o)

	o.Name = "user1_updated"
	o.Params.TelegramID = 1234

	err = db.UsersUpdate(o)
	require.NoError(t, err)

	o2, err := db.UsersGetByID(1)
	require.NoError(t, err)
	assert.Equal(t, "user1_updated", o2.Name)
	assert.Equal(t, o.Params, o2.Params)
}

func TestUsersGetByParams_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByParam(models.UserTelegramID, 31337)
	assert.NoError(t, err)
	assert.NotNil(t, o)
	assert.Equal(t, "user1", o.Name)
}

func TestUsersGetByParams_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByParam(models.UserTelegramID, 1)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}
