package database_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/russtone/sonar/internal/database/models"
)

func TestPayloadsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.Payload{
		UserID:          1,
		Subdomain:       "8a8b58beaf",
		Name:            "test",
		NotifyProtocols: models.ProtoCategoriesAll,
		StoreEvents:     true,
	}

	err := db.PayloadsCreate(o)
	assert.NoError(t, err)
	assert.NotZero(t, o.ID)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)
	assert.Equal(t, models.ProtoCategoriesAll, o.NotifyProtocols)
	assert.Equal(t, true, o.StoreEvents)
}

func TestPayloadsCreate_Duplicate(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.Payload{
		UserID:    1,
		Subdomain: "8a8b58beaf",
		Name:      "payload1",
	}

	err := db.PayloadsCreate(o)
	assert.Error(t, err)
}

func TestPayloadGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadGetByID(1)
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.Equal(t, "payload1", o.Name)
}

func TestPayloadGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadGetByID(1337)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.Error(t, err, sql.ErrNoRows.Error())
}

func TestPayloadsGetBySubdomain_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadsGetBySubdomain("c1da9f3d")
	assert.NoError(t, err)
	assert.Equal(t, "payload1", o.Name)
}

func TestPayloadsGetBySubdomain_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsGetBySubdomain("not_exist")
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestPayloadsGetByUserAndName_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadsGetByUserAndName(1, "payload1")
	assert.NoError(t, err)
	assert.Equal(t, "c1da9f3d", o.Subdomain)
}

func TestPayloadsGetByUserAndName_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsGetByUserAndName(1, "not_exist")
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestPayloadsFindByUserID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	pp, err := db.PayloadsFindByUserID(2)
	assert.NoError(t, err)
	assert.Len(t, pp, 2)

	names := make([]string, 0)
	for _, p := range pp {
		names = append(names, p.Name)
	}

	assert.Contains(t, names, "payload2")
	assert.Contains(t, names, "payload3")
}

func TestPayloadsFindByUserID_Empty(t *testing.T) {
	setup(t)
	defer teardown(t)

	pp, err := db.PayloadsFindByUserID(3)
	assert.NoError(t, err)
	assert.Len(t, pp, 0)
}

func TestPayloadsFindByUserAndName_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	pp, err := db.PayloadsFindByUserAndName(1, "payload1")
	assert.NoError(t, err)
	assert.Len(t, pp, 1)

	subdomains := make([]string, 0)
	for _, p := range pp {
		subdomains = append(subdomains, p.Subdomain)
	}

	assert.Contains(t, subdomains, "c1da9f3d")
}

func TestPayloadsFindByUserAndName_Empty(t *testing.T) {
	setup(t)
	defer teardown(t)

	pp, err := db.PayloadsFindByUserAndName(3, "payload1")
	assert.NoError(t, err)
	assert.Len(t, pp, 0)
}

func TestPayloadsDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.PayloadsDelete(1)
	assert.NoError(t, err)
}

func TestPayloadsDelete_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.PayloadsDelete(1337)
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestPayloadsUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadGetByID(1)
	require.NoError(t, err)
	assert.NotNil(t, o)

	o.Name = "payload1_updated"
	o.NotifyProtocols = models.ProtoCategories("dns")
	o.StoreEvents = false

	err = db.PayloadsUpdate(o)
	require.NoError(t, err)

	o2, err := db.PayloadGetByID(1)
	require.NoError(t, err)
	assert.Equal(t, o, o2)
}

func TestPayloadsDeleteByNamePart_All_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	removed, err := db.PayloadsDeleteByNamePart(1, "")
	require.NoError(t, err)
	require.Len(t, removed, 2)

	left, err := db.PayloadsFindByUserID(1)
	require.NoError(t, err)
	require.Len(t, left, 0)
}

func TestPayloadsDeleteByNamePart_Substr_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	removed, err := db.PayloadsDeleteByNamePart(1, "1")
	require.NoError(t, err)
	require.Len(t, removed, 1)

	left, err := db.PayloadsFindByUserID(1)
	require.NoError(t, err)
	require.Len(t, left, 1)
}
