package cmd_test

import (
	"context"
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePayload_Success(t *testing.T) {
	c, acts, hnd := prepare()

	res := actions.PayloadsCreateResult(&actions.Payload{})

	acts.
		On("PayloadsCreate", ctx, actions.PayloadsCreateParams{
			Name:            "test",
			NotifyProtocols: models.PayloadProtocolsAll,
		}).
		Return(res, nil)

	hnd.
		On("PayloadsCreate", ctx, res)

	_, err := c.Exec(ctx, &actions.User{}, []string{"new", "test"})
	assert.NoError(t, err)

	acts.AssertExpectations(t)
	hnd.AssertExpectations(t)
}

func TestCreatePayload_NoArg(t *testing.T) {
	c, _, _ := prepare()

	out, err := c.Exec(context.Background(), &actions.User{}, []string{"new"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestDeletePayload_Success(t *testing.T) {
	c, acts, hnd := prepare()

	res := actions.PayloadsDeleteResult(&actions.Payload{})

	acts.
		On("PayloadsDelete", ctx, actions.PayloadsDeleteParams{
			Name: "test",
		}).
		Return(res, nil)

	hnd.
		On("PayloadsDelete", ctx, res)

	_, err := c.Exec(ctx, &actions.User{}, []string{"del", "test"})
	assert.NoError(t, err)

	acts.AssertExpectations(t)
	hnd.AssertExpectations(t)
}

func TestDeletePayload_NoArg(t *testing.T) {
	c, _, _ := prepare()

	out, err := c.Exec(context.Background(), &actions.User{}, []string{"del"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestListPayloads_Success(t *testing.T) {
	c, acts, hnd := prepare()

	res := actions.PayloadsListResult([]*actions.Payload{})

	acts.
		On("PayloadsList", ctx, actions.PayloadsListParams{
			Name: "test",
		}).
		Return(res, nil)

	hnd.
		On("PayloadsList", ctx, res)

	_, err := c.Exec(ctx, &actions.User{}, []string{"list", "test"})
	assert.NoError(t, err)

	acts.AssertExpectations(t)
	hnd.AssertExpectations(t)
}

func TestUpdatePayload_Success(t *testing.T) {
	c, acts, hnd := prepare()

	res := actions.PayloadsUpdateResult(&actions.Payload{})

	acts.
		On("PayloadsUpdate", ctx, actions.PayloadsUpdateParams{
			Name:            "payload1",
			NewName:         "payload1_updated",
			NotifyProtocols: []string{"dns", "http"},
		}).
		Return(res, nil)

	hnd.
		On("PayloadsUpdate", ctx, res)

	_, err := c.Exec(ctx, &actions.User{},
		[]string{"mod", "payload1", "-n", "payload1_updated", "-p", "dns,http"})
	assert.NoError(t, err)

	acts.AssertExpectations(t)
	hnd.AssertExpectations(t)
}
