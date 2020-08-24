package dbactions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func TestCreatePayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		p    actions.CreatePayloadParams
	}{
		{
			"empty notify protocols",
			actions.CreatePayloadParams{
				Name:            "test",
				NotifyProtocols: []string{},
			},
		},
		{
			"dns only",
			actions.CreatePayloadParams{
				Name:            "test-dns",
				NotifyProtocols: []string{models.PayloadProtocolDNS},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.CreatePayload(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
			assert.Equal(t, tt.p.Name, r.Name)
			assert.Equal(t, tt.p.NotifyProtocols, []string(r.NotifyProtocols))
			assert.Equal(t, u.ID, r.UserID)
		})
	}
}

func TestCreatePayload_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.CreatePayloadParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.CreatePayloadParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			ctx,
			actions.CreatePayloadParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"duplicate payload name",
			ctx,
			actions.CreatePayloadParams{
				Name: "payload1",
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.CreatePayload(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDeletePayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		p    actions.DeletePayloadParams
	}{
		{
			"payload1",
			actions.DeletePayloadParams{
				Name: "payload1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.DeletePayload(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
		})
	}
}

func TestDeletePayload_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.DeletePayloadParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.DeletePayloadParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			ctx,
			actions.DeletePayloadParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload name",
			ctx,
			actions.DeletePayloadParams{
				Name: "payload2",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DeletePayload(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestListPayloads_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name  string
		p     actions.ListPayloadsParams
		count int
	}{
		{
			"all",
			actions.ListPayloadsParams{
				Name: "",
			},
			2,
		},
		{
			"payload1",
			actions.ListPayloadsParams{
				Name: "payload1",
			},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.ListPayloads(ctx, tt.p)
			assert.NoError(t, err)
			assert.Len(t, r, tt.count)
		})
	}
}

func TestListPayloads_Error(t *testing.T) {

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.ListPayloadsParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.ListPayloadsParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.ListPayloads(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestUpdatePayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		p    actions.UpdatePayloadParams
	}{
		{
			"update name",
			actions.UpdatePayloadParams{
				Name:    "payload1",
				NewName: "payload1_updated",
			},
		},
		{
			"update notify protocols",
			actions.UpdatePayloadParams{
				Name:            "payload1",
				NotifyProtocols: []string{models.PayloadProtocolHTTP},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.UpdatePayload(ctx, tt.p)
			assert.NoError(t, err)
			assert.NotNil(t, r)

			if tt.p.NewName != "" {
				assert.Equal(t, tt.p.NewName, r.Name)
			}

			if tt.p.NotifyProtocols != nil {
				assert.Equal(t, tt.p.NotifyProtocols, []string(r.NotifyProtocols))
			}
		})
	}
}

func TestUpdatePayload_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actions.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.UpdatePayloadParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.UpdatePayloadParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty name",
			ctx,
			actions.UpdatePayloadParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			ctx,
			actions.UpdatePayloadParams{
				Name: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.UpdatePayload(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
