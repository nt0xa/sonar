package actionsdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func TestAuditWrite_OnPayloadCreate(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)
	actsAudit := actionsdb.New(db, log, "sonar.test", true)

	_, err = actsAudit.PayloadsCreate(ctx, actions.PayloadsCreateParams{
		Name:            "audit-write-test",
		NotifyProtocols: []string{database.ProtoCategoryDNS},
	})
	require.NoError(t, err)

	recs, err := db.AuditRecordsList(t.Context(), database.AuditRecordsListParams{
		ActorID:      nil,
		ResourceType: "payload",
		Action:       "create",
		ActorName:    "",
		FromAt:       nil,
		ToAt:         nil,
		PageLimit:    10,
		PageOffset:   0,
	})
	require.NoError(t, err)
	require.NotEmpty(t, recs)
	assert.Equal(t, "payload", string(recs[0].ResourceType))
	assert.Equal(t, "create", string(recs[0].Action))
	assert.Equal(t, "audit-write-test", recs[0].Resource["name"])
	require.NotNil(t, recs[0].ActorID)
	assert.EqualValues(t, 1, *recs[0].ActorID)
}

func TestAuditRecordsListGet_AdminOnly(t *testing.T) {
	setup(t)
	defer teardown(t)

	actsAudit := actionsdb.New(db, log, "sonar.test", true)

	// Seed one audit row
	admin, err := db.UsersGetByID(t.Context(), 3)
	require.NoError(t, err)
	adminCtx := actionsdb.SetUser(t.Context(), admin)

	_, err = actsAudit.PayloadsCreate(adminCtx, actions.PayloadsCreateParams{
		Name: "audit-admin-list",
	})
	require.NoError(t, err)

	list, err := actsAudit.AuditRecordsList(adminCtx, actions.AuditRecordsListParams{PerPage: 5})
	require.NoError(t, err)
	require.NotEmpty(t, list)

	got, err := actsAudit.AuditRecordsGet(adminCtx, actions.AuditRecordsGetParams{ID: list[0].ID})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, list[0].ID, got.ID)

	nonAdmin, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)
	nonAdminCtx := actionsdb.SetUser(t.Context(), nonAdmin)

	_, err = actsAudit.AuditRecordsList(nonAdminCtx, actions.AuditRecordsListParams{PerPage: 5})
	assert.Error(t, err)
	assert.IsType(t, &errors.ForbiddenError{}, err)
}
