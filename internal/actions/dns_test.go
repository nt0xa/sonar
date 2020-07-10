package actions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
)

func TestCreateDNSRecord_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	p := actions.CreateDNSRecordParams{
		Name:        "test",
		PayloadName: "payload1",
		TTL:         60,
		Type:        models.DNSTypeA,
		Strategy:    models.DNSStrategyAll,
		Values:      []string{"127.0.0.1"},
	}

	r, err := acts.CreateDNSRecord(u, p)
	require.NoError(t, err)
	require.NotNil(t, r)

	assert.Equal(t, "test", r.Record.Name)
}

func TestDeleteDNSRecord_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	p := actions.DeleteDNSRecordParams{
		Name:        "test-a",
		PayloadName: "payload1",
		Type:        models.DNSTypeA,
	}

	_, err = acts.DeleteDNSRecord(u, p)
	require.NoError(t, err)
}

func TestListDNSRecords_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	p := actions.ListDNSRecordsParams{
		PayloadName: "payload1",
	}

	r, err := acts.ListDNSRecords(u, p)
	require.NoError(t, err)
	require.Len(t, r.Records, 9)
}
