package database_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database"
)

func TestAuditRecordsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	actorID := int64(2)
	rec, err := db.AuditRecordsCreate(t.Context(), database.AuditRecordsCreateParams{
		Action:       database.AuditRecordActionTypeCreate,
		ResourceType: database.AuditRecordResourceTypeDNSRecord,
		Source:       database.AuditRecordSourceTypeLark,
		ActorID:      &actorID,
		ActorName:    "user2",
		ActorMetadata: database.AuditActorMetadata{
			"lark_id": "8b2494d",
		},
		Resource: database.AuditResource{
			"name":  "a",
			"type":  "A",
			"value": "127.0.0.1",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.NotZero(t, rec.ID)
	assert.NotEqual(t, uuid.Nil, rec.UUID)
	assert.WithinDuration(t, time.Now(), rec.CreatedAt, 5*time.Second)
	assert.Equal(t, database.AuditRecordActionTypeCreate, rec.Action)
	assert.Equal(t, database.AuditRecordResourceTypeDNSRecord, rec.ResourceType)
	assert.Equal(t, database.AuditRecordSourceTypeLark, rec.Source)
	require.NotNil(t, rec.ActorID)
	assert.EqualValues(t, 2, *rec.ActorID)
	assert.Equal(t, "user2", rec.ActorName)
	assert.EqualValues(t, "8b2494d", rec.ActorMetadata["lark_id"])
	assert.EqualValues(t, "a", rec.Resource["name"])
}

func TestAuditRecordsGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	rec, err := db.AuditRecordsGetByID(t.Context(), 1)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.EqualValues(t, 1, rec.ID)
	assert.Equal(t, database.AuditRecordActionTypeCreate, rec.Action)
	assert.Equal(t, database.AuditRecordResourceTypePayload, rec.ResourceType)
	assert.Equal(t, database.AuditRecordSourceTypeAPI, rec.Source)
	require.NotNil(t, rec.ActorID)
	assert.EqualValues(t, 1, *rec.ActorID)
	assert.Equal(t, "user1", rec.ActorName)
	assert.EqualValues(t, "payload1", rec.Resource["name"])
}

func TestAuditRecordsGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.AuditRecordsGetByID(t.Context(), 1337)
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestAuditRecordsList_FilterAndPagination(t *testing.T) {
	setup(t)
	defer teardown(t)

	// No filters, first page.
	all, err := db.AuditRecordsList(t.Context(), database.AuditRecordsListParams{
		PageLimit:  2,
		PageOffset: 0,
	})
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.EqualValues(t, 3, all[0].ID)
	assert.EqualValues(t, 2, all[1].ID)

	// Pagination.
	next, err := db.AuditRecordsList(t.Context(), database.AuditRecordsListParams{
		PageLimit:  2,
		PageOffset: 2,
	})
	require.NoError(t, err)
	require.Len(t, next, 1)
	assert.EqualValues(t, 1, next[0].ID)

	// By actor.
	actorID := int64(1)
	byActor, err := db.AuditRecordsList(t.Context(), database.AuditRecordsListParams{
		ActorID:    &actorID,
		PageLimit:  10,
		PageOffset: 0,
	})
	require.NoError(t, err)
	require.Len(t, byActor, 2)
	assert.EqualValues(t, 2, byActor[0].ID)
	assert.EqualValues(t, 1, byActor[1].ID)

	// By action/resource type.
	byActionType, err := db.AuditRecordsList(t.Context(), database.AuditRecordsListParams{
		Action:       "delete",
		ResourceType: "user",
		PageLimit:    10,
		PageOffset:   0,
	})
	require.NoError(t, err)
	require.Len(t, byActionType, 1)
	assert.EqualValues(t, 3, byActionType[0].ID)

	// By time range.
	from := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 2, 23, 59, 59, 0, time.UTC)
	byTime, err := db.AuditRecordsList(t.Context(), database.AuditRecordsListParams{
		FromAt:       &from,
		ToAt:         &to,
		PageLimit:    10,
		PageOffset:   0,
		ActorName:    "",
		Action:       "",
		ResourceType: "",
	})
	require.NoError(t, err)
	require.Len(t, byTime, 1)
	assert.EqualValues(t, 3, byTime[0].ID)
}
