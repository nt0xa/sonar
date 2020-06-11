package cmd_test

import (
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	_, err := execute(cmd, user, "new test")
	assert.NoError(t, err)
}

func TestCreatePayload_NoArg(t *testing.T) {
	cmd, _, _ := prepare()

	out, err := execute(cmd, user, "new")
	assert.Error(t, err)
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

	_, err := execute(cmd, user, "del test")
	assert.NoError(t, err)
}

func TestDeletePayload_NoArg(t *testing.T) {
	cmd, _, _ := prepare()

	out, err := execute(cmd, user, "del")
	assert.Error(t, err)
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

	_, err := execute(cmd, user, "list test")
	assert.NoError(t, err)
}
