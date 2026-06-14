package auditsvc_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/internal/service/auditsvc"
)

// user1's API token, see fixtures/users.yml.
const user1Token = "50c862e41d059eeca13adc7b276b46b7"

func TestAudit_OnPayloadCreate(t *testing.T) {
	setup(t)
	defer teardown(t)

	ctx, err := svc.AuthContextByAPIToken(t.Context(), user1Token)
	require.NoError(t, err)

	_, err = svc.PayloadsCreate(ctx, service.PayloadsCreateInput{Name: "audit-test"})
	require.NoError(t, err)

	// Audit records are written in the background; wait for them to flush.
	svc.(*auditsvc.Service).Wait()

	recs, err := db.AuditRecordsList(t.Context(), database.AuditRecordsListParams{
		ResourceType: "payload",
		Action:       "create",
		ActorName:    "user1",
		PageLimit:    10,
	})
	require.NoError(t, err)
	require.NotEmpty(t, recs)

	rec := recs[0]
	assert.Equal(t, "payload", string(rec.ResourceType))
	assert.Equal(t, "create", string(rec.Action))
	assert.Equal(t, "api", string(rec.Source))
	require.NotNil(t, rec.ActorID)
	assert.EqualValues(t, 1, *rec.ActorID)
	assert.Equal(t, "user1", rec.ActorName)

	var resource map[string]any
	require.NoError(t, json.Unmarshal(rec.Resource, &resource))
	assert.Equal(t, "audit-test", resource["name"])
}
