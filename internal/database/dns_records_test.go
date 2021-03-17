package database_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/models"
)

func TestDNSRecordsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.DNSRecord{
		PayloadID: 1,
		Name:      "test",
		Type:      models.DNSTypeA,
		TTL:       60,
		Values:    []string{"127.0.0.1"},
		Strategy:  models.DNSStrategyAll,
	}

	err := db.DNSRecordsCreate(o)
	assert.NoError(t, err)
	assert.NotZero(t, o.ID)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)
}

func TestDNSRecordsCreate_Duplicate(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.DNSRecord{
		PayloadID: 1,
		Name:      "dns1",
		Type:      models.DNSTypeA,
		TTL:       60,
		Values:    []string{"127.0.0.1"},
	}

	err := db.DNSRecordsCreate(o)
	assert.Error(t, err)
}

func TestDNSRecordsGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "test-a", o.Name)
}

func TestDNSRecordsGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByID(1337)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.Error(t, err, sql.ErrNoRows.Error())
}

func TestDNSRecordsGetByPayloadNameType_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByPayloadNameType(1, "test-a", models.DNSTypeA)
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.Equal(t, int64(1), o.ID)
}

func TestDNSRecordsGetByPayloadNameType_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByPayloadNameType(1337, "dns1", models.DNSTypeA)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.Error(t, err, sql.ErrNoRows.Error())
}

func TestDNSRecordsDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.DNSRecordsDelete(1)
	assert.NoError(t, err)
}

func TestDNSRecordsUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByID(1)
	require.NoError(t, err)
	assert.NotNil(t, o)

	o.Name = "dns1-updated"
	o.Values = []string{"127.0.0.1", "127.0.0.2"}

	err = db.DNSRecordsUpdate(o)
	require.NoError(t, err)

	o2, err := db.DNSRecordsGetByID(1)
	require.NoError(t, err)
	assert.Equal(t, o, o2)
}
