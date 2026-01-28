package database2_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database2"
)

func ptr[T any](v T) *T {
	return &v
}

func TestDNSRecordsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsCreate(t.Context(), database2.DNSRecordsCreateParams{
		PayloadID: 1,
		Name:      "test",
		Type:      database2.DNSRecordTypeA,
		TTL:       60,
		Values:    []string{"127.0.0.1"},
		Strategy:  database2.DNSStrategyAll,
	})
	assert.NoError(t, err)
	assert.NotNil(t, o.ID)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)
	assert.EqualValues(t, 10, o.Index)
}

func TestDNSRecordsCreate_Duplicate(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.DNSRecordsCreate(t.Context(), database2.DNSRecordsCreateParams{
		PayloadID: 1,
		Name:      "test-a",
		Type:      database2.DNSRecordTypeA,
		TTL:       60,
		Values:    []string{"127.0.0.1"},
	})
	assert.Error(t, err)
}

func TestDNSRecordsGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByID(t.Context(), ptr(int64(1)))
	assert.NoError(t, err)
	assert.Equal(t, "test-a", o.Name)
}

func TestDNSRecordsGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.DNSRecordsGetByID(t.Context(), ptr(int64(1337)))
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestDNSRecordsGetByPayloadNameAndType_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByPayloadNameAndType(t.Context(), database2.DNSRecordsGetByPayloadNameAndTypeParams{
		PayloadID: 1,
		Name:      "test-a",
		Type:      database2.DNSRecordTypeA,
	})
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.Equal(t, ptr(int64(1)), o.ID)
}

func TestDNSRecordsGetByPayloadNameAndType_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.DNSRecordsGetByPayloadNameAndType(t.Context(), database2.DNSRecordsGetByPayloadNameAndTypeParams{
		PayloadID: 1337,
		Name:      "dns1",
		Type:      database2.DNSRecordTypeA,
	})
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestDNSRecordsDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	err := db.DNSRecordsDelete(t.Context(), ptr(int64(1)))
	assert.NoError(t, err)
}

func TestDNSRecordsUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByID(t.Context(), ptr(int64(1)))
	require.NoError(t, err)
	assert.NotNil(t, o)

	updated, err := db.DNSRecordsUpdate(t.Context(), database2.DNSRecordsUpdateParams{
		ID:             o.ID,
		PayloadID:      o.PayloadID,
		Name:           "dns1-updated",
		Type:           o.Type,
		TTL:            o.TTL,
		Values:         []string{"127.0.0.1", "127.0.0.2"},
		Strategy:       o.Strategy,
		LastAnswer:     o.LastAnswer,
		LastAccessedAt: o.LastAccessedAt,
	})
	require.NoError(t, err)

	o2, err := db.DNSRecordsGetByID(t.Context(), ptr(int64(1)))
	require.NoError(t, err)
	assert.Equal(t, updated.Name, o2.Name)
	assert.Equal(t, updated.Values, o2.Values)
}

func TestDNSRecordsGetByPayloadID(t *testing.T) {
	setup(t)
	defer teardown(t)

	l, err := db.DNSRecordsGetByPayloadID(t.Context(), 1)
	assert.NoError(t, err)
	assert.Len(t, l, 9)
	assert.EqualValues(t, 1, l[0].Index)
	assert.EqualValues(t, 9, l[len(l)-1].Index)
}

func TestDNSRecordsGetCountByPayloadID(t *testing.T) {
	setup(t)
	defer teardown(t)

	res, err := db.DNSRecordsGetCountByPayloadID(t.Context(), 1)
	assert.NoError(t, err)
	assert.EqualValues(t, 9, res)

	res, err = db.DNSRecordsGetCountByPayloadID(t.Context(), 2)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, res)
}

func TestDNSRecordsGetByPayloadIDAndIndex(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.DNSRecordsGetByPayloadIDAndIndex(t.Context(), 1, 2)
	assert.NoError(t, err)
	assert.EqualValues(t, "test-aaaa", o.Name)

	// Not exist
	_, err = db.DNSRecordsGetByPayloadIDAndIndex(t.Context(), 1, 1337)
	assert.Error(t, err)
}

func TestDNSRecordsDeleteAllByPayloadID(t *testing.T) {
	setup(t)
	defer teardown(t)

	l, err := db.DNSRecordsDeleteAllByPayloadID(t.Context(), 1)
	assert.NoError(t, err)
	assert.Len(t, l, 9)
}

func TestDNSRecordsDeleteAllByPayloadIDAndName(t *testing.T) {
	setup(t)
	defer teardown(t)

	l, err := db.DNSRecordsDeleteAllByPayloadIDAndName(t.Context(), 1, "test-a")
	assert.NoError(t, err)
	assert.Len(t, l, 1)
}
