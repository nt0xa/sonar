package dbactions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func TestCreatePayload_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.CreatePayloadParams{
		Name: "test",
	}

	r, err := acts.CreatePayload(ctx, p)
	assert.NoError(t, err)

	assert.NotNil(t, r)
	assert.Equal(t, "test", r.Name)
	assert.Equal(t, u.ID, r.UserID)
}

func TestCreatePayload_Validation(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.CreatePayloadParams{
		Name: "",
	}

	_, err = acts.CreatePayload(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.ValidationError{}, err)
}

func TestCreatePayload_Conflict(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.CreatePayloadParams{
		Name: "payload1",
	}

	_, err = acts.CreatePayload(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.ConflictError{}, err)
}

func TestDeletePayload_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.DeletePayloadParams{
		Name: "payload1",
	}

	_, err = acts.DeletePayload(ctx, p)
	assert.NoError(t, err)
}

func TestDeletePayload_Validation(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.DeletePayloadParams{
		Name: "",
	}

	_, err = acts.DeletePayload(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.ValidationError{}, err)
}

func TestDeletePayload_NotFound(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.DeletePayloadParams{
		Name: "not-exist",
	}

	_, err = acts.DeletePayload(ctx, p)
	assert.Error(t, err)
	assert.IsType(t, &errors.NotFoundError{}, err)
}

func TestListPayloads_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(2)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.ListPayloadsParams{
		Name: "",
	}

	r, err := acts.ListPayloads(ctx, p)
	assert.NoError(t, err)
	assert.Len(t, r, 2)

	p = actions.ListPayloadsParams{
		Name: "payload2",
	}

	r, err = acts.ListPayloads(ctx, p)
	assert.NoError(t, err)
	assert.Len(t, r, 1)
}

func TestUpdatePayload_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	p := actions.UpdatePayloadParams{
		Name:            "payload1",
		NewName:         "payload1_updated",
		NotifyProtocols: []string{"dns"},
	}

	_, err = acts.UpdatePayload(ctx, p)
	assert.NoError(t, err)

	p2, err := db.PayloadsGetByUserAndName(1, "payload1_updated")
	require.NoError(t, err)
	require.NotNil(t, p2)
	assert.Equal(t, "payload1_updated", p2.Name)
	assert.Equal(t, []string{"dns"}, []string(p2.NotifyProtocols))
}
