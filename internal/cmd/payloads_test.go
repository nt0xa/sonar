package cmd_test

import (
	"context"
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreatePayload_Success(t *testing.T) {
	cmd, acts, hnd := prepare()

	res := payload

	acts.
		On("CreatePayload", user, actions.CreatePayloadParams{
			Name: "test",
		}).
		Return(res, nil)

	hnd.
		On("Handle", mock.Anything, res)

	_, err := cmd.Exec(context.Background(), user, []string{"new", "test"})
	assert.NoError(t, err)
}

func TestCreatePayload_NoArg(t *testing.T) {
	cmd, _, _ := prepare()

	out, err := cmd.Exec(context.Background(), user, []string{"new"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestDeletePayload_Success(t *testing.T) {
	cmd, acts, hnd := prepare()

	res := &actions.MessageResult{Message: "test"}

	acts.
		On("DeletePayload", user, actions.DeletePayloadParams{
			Name: "test",
		}).
		Return(res, nil)

	hnd.
		On("Handle", mock.Anything, res)

	_, err := cmd.Exec(context.Background(), user, []string{"del", "test"})
	assert.NoError(t, err)
}

func TestDeletePayload_NoArg(t *testing.T) {
	cmd, _, _ := prepare()

	out, err := cmd.Exec(context.Background(), user, []string{"del"})
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestListPayloads_Success(t *testing.T) {
	cmd, acts, hnd := prepare()

	res := payloads

	acts.
		On("ListPayloads", user, actions.ListPayloadsParams{
			Name: "test",
		}).
		Return(res, nil)

	hnd.
		On("Handle", mock.Anything, res)

	_, err := cmd.Exec(context.Background(), user, []string{"list", "test"})
	assert.NoError(t, err)
}
