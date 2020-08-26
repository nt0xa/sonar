package apiclient_test

import (
	"context"
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSRecordsCreate(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.DNSRecordsCreateParams{
		PayloadName: "payload1",
		Name:        "test",
		TTL:         60,
		Type:        models.DNSTypeA,
		Values:      []string{"1.1.1.1", "2.2.2.2"},
		Strategy:    models.DNSStrategyAll,
	}

	res, err := client.DNSRecordsCreate(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, p.Name, res.Record.Name)
}

func TestDNSRecordsList(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.DNSRecordsListParams{
		PayloadName: "payload1",
	}

	res, err := client.DNSRecordsList(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Len(t, res.Records, 9)
}

func TestDNSRecordsDelete(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.DNSRecordsDeleteParams{
		PayloadName: "payload1",
		Name:        "test-aaaa",
		Type:        models.DNSTypeAAAA,
	}

	res, err := client.DNSRecordsDelete(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)
}
