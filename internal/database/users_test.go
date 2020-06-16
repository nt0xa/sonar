package database_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUsersCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.User{
		ID:   1337,
		Name: "test",
		Params: models.UserParams{
			TelegramID: 31337,
		},
	}

	err := db.UsersCreate(o)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now().UTC(), o.CreatedAt, 5*time.Second)
}

func TestUsersCreate_Duplicate(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.User{
		ID:   1337,
		Name: "user1",
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

func TestUsersDelete_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.UsersDelete(1337)
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestUsersUpdate_Succes(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.UsersGetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, o)

	o.Name = "user1_updated"

	err = db.UsersUpdate(o)
	assert.NoError(t, err)

	o2, err := db.UsersGetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "user1_updated", o2.Name)
}

func TestUsersUpdate_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.User{ID: 1337}

	err := db.UsersUpdate(o)
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestUsersGetByParams_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := &models.UserParams{TelegramID: 31337}

	o, err := db.UsersGetByParams(p)
	assert.NoError(t, err)
	assert.NotNil(t, o)
	assert.Equal(t, "user1", o.Name)
}

func TestUsersGetByParams_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := &models.UserParams{TelegramID: 1}

	o, err := db.UsersGetByParams(p)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}
