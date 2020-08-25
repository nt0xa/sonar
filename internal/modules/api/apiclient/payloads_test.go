package apiclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
)

func TestPayloadCreate(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.PayloadsCreateParams{
		Name:            "test",
		NotifyProtocols: []string{"http"},
	}

	res, err := client.PayloadsCreate(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, p.Name, res.Name)
	assert.Equal(t, p.NotifyProtocols, res.NotifyProtocols)
}

func TestPayloadList(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.PayloadsListParams{
		Name: "",
	}

	res, err := client.PayloadsList(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Len(t, res, 2)
}

func TestPayloadDelete(t *testing.T) {
	setup(t)
	defer teardown(t)

	p := actions.PayloadsDeleteParams{
		Name: "payload1",
	}

	res, err := client.PayloadsDelete(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, res)
}
