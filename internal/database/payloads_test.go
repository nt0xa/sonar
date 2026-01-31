package database_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database"
)

func TestPayloadsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadsCreate(t.Context(), database.PayloadsCreateParams{
		UserID:          1,
		Subdomain:       "8a8b58beaf",
		Name:            "test",
		NotifyProtocols: []string{"dns", "http", "smtp"},
		StoreEvents:     true,
	})
	assert.NoError(t, err)
	assert.NotZero(t, o.ID)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)
	assert.Equal(t, []string{"dns", "http", "smtp"}, o.NotifyProtocols)
	assert.Equal(t, true, o.StoreEvents)
}

func TestPayloadsCreate_Duplicate(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsCreate(t.Context(), database.PayloadsCreateParams{
		UserID:    1,
		Subdomain: "8a8b58beaf",
		Name:      "payload1",
	})
	assert.Error(t, err)
}

func TestPayloadsGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadsGetByID(t.Context(), 1)
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.Equal(t, "payload1", o.Name)
}

func TestPayloadsGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsGetByID(t.Context(), 1337)
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestPayloadsGetBySubdomain_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadsGetBySubdomain(t.Context(), "c1da9f3d")
	assert.NoError(t, err)
	assert.Equal(t, "payload1", o.Name)
}

func TestPayloadsGetBySubdomain_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsGetBySubdomain(t.Context(), "not_exist")
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestPayloadsGetByUserAndName_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadsGetByUserAndName(t.Context(), 1, "payload1")
	assert.NoError(t, err)
	assert.Equal(t, "c1da9f3d", o.Subdomain)
}

func TestPayloadsGetByUserAndName_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsGetByUserAndName(t.Context(), 1, "not_exist")
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestPayloadsFindByUserID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	pp, err := db.PayloadsFindByUserID(t.Context(), 2)
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

	pp, err := db.PayloadsFindByUserID(t.Context(), 3)
	assert.NoError(t, err)
	assert.Len(t, pp, 0)
}

func TestPayloadsFindByUserAndName_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	pp, err := db.PayloadsFindByUserAndName(t.Context(), database.PayloadsFindByUserAndNameParams{
		UserID: 1,
		Name:   "%payload1%",
		Limit:  10,
		Offset: 0,
	})
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

	pp, err := db.PayloadsFindByUserAndName(t.Context(), database.PayloadsFindByUserAndNameParams{
		UserID: 3,
		Name:   "%payload1%",
		Limit:  10,
		Offset: 0,
	})
	assert.NoError(t, err)
	assert.Len(t, pp, 0)
}

func TestPayloadsFindByUserAndName_Pagination(t *testing.T) {
	setup(t)
	defer teardown(t)

	pp, err := db.PayloadsFindByUserAndName(t.Context(), database.PayloadsFindByUserAndNameParams{
		UserID: 1,
		Name:   "%%",
		Limit:  3,
		Offset: 0,
	})
	assert.NoError(t, err)
	assert.Len(t, pp, 3)

	assert.EqualValues(t, 9, pp[0].ID)
	assert.EqualValues(t, 7, pp[len(pp)-1].ID)

	pp, err = db.PayloadsFindByUserAndName(t.Context(), database.PayloadsFindByUserAndNameParams{
		UserID: 1,
		Name:   "%%",
		Limit:  3,
		Offset: 3,
	})
	assert.NoError(t, err)
	assert.Len(t, pp, 3)

	assert.EqualValues(t, 6, pp[0].ID)
	assert.EqualValues(t, 1, pp[len(pp)-1].ID)
}

func TestPayloadsDelete_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsDelete(t.Context(), 1)
	assert.NoError(t, err)
}

func TestPayloadsDelete_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.PayloadsDelete(t.Context(), 1337)
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestPayloadsUpdate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.PayloadsGetByID(t.Context(), 1)
	require.NoError(t, err)
	assert.NotNil(t, o)

	updated, err := db.PayloadsUpdate(t.Context(), database.PayloadsUpdateParams{
		ID:              o.ID,
		UserID:          o.UserID,
		Subdomain:       o.Subdomain,
		Name:            "payload1_updated",
		NotifyProtocols: []string{"dns"},
		StoreEvents:     false,
	})
	require.NoError(t, err)

	o2, err := db.PayloadsGetByID(t.Context(), 1)
	require.NoError(t, err)
	assert.Equal(t, updated.Name, o2.Name)
	assert.Equal(t, updated.NotifyProtocols, o2.NotifyProtocols)
	assert.Equal(t, updated.StoreEvents, o2.StoreEvents)
}

func TestPayloadsDeleteByNamePart_All_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	removed, err := db.PayloadsDeleteByNamePart(t.Context(), 1, "%%")
	require.NoError(t, err)
	require.Len(t, removed, 6)

	left, err := db.PayloadsFindByUserID(t.Context(), 1)
	require.NoError(t, err)
	require.Len(t, left, 0)
}

func TestPayloadsDeleteByNamePart_Substr_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	removed, err := db.PayloadsDeleteByNamePart(t.Context(), 1, "%1%")
	require.NoError(t, err)
	require.Len(t, removed, 2)

	left, err := db.PayloadsFindByUserID(t.Context(), 1)
	require.NoError(t, err)
	require.Len(t, left, 4)
}
