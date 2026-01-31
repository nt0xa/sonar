package actionsdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/errors"
	"github.com/nt0xa/sonar/internal/utils/pointer"
)

func TestCreatePayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		p    actions.PayloadsCreateParams
	}{
		{
			"empty notify protocols",
			actions.PayloadsCreateParams{
				Name:            "test",
				NotifyProtocols: []string{},
			},
		},
		{
			"dns only",
			actions.PayloadsCreateParams{
				Name:            "test-dns",
				NotifyProtocols: []string{database.ProtoCategoryDNS},
			},
		},
		{
			"store events",
			actions.PayloadsCreateParams{
				Name:            "test-dns",
				StoreEvents:     true,
				NotifyProtocols: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.PayloadsCreate(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
			assert.Equal(t, tt.p.Name, r.Name)
			assert.Equal(t, tt.p.NotifyProtocols, r.NotifyProtocols)
			assert.Equal(t, tt.p.StoreEvents, r.StoreEvents)
		})
	}
}

func TestCreatePayload_Error(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsCreateParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			t.Context(),
			actions.PayloadsCreateParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			ctx,
			actions.PayloadsCreateParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"duplicate payload name",
			ctx,
			actions.PayloadsCreateParams{
				Name: "payload1",
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.PayloadsCreate(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDeletePayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		p    actions.PayloadsDeleteParams
	}{
		{
			"payload1",
			actions.PayloadsDeleteParams{
				Name: "payload1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.PayloadsDelete(ctx, tt.p)
			assert.NoError(t, err)

			assert.NotNil(t, r)
		})
	}
}

func TestDeletePayload_Error(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsDeleteParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			t.Context(),
			actions.PayloadsDeleteParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			ctx,
			actions.PayloadsDeleteParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload name",
			ctx,
			actions.PayloadsDeleteParams{
				Name: "payload2",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.PayloadsDelete(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestListPayloads_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name  string
		p     actions.PayloadsListParams
		count int
	}{
		{
			"all",
			actions.PayloadsListParams{
				Name: "",
			},
			6,
		},
		{
			"payload1",
			actions.PayloadsListParams{
				Name: "payload1",
			},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.PayloadsList(ctx, tt.p)
			assert.NoError(t, err)
			assert.Len(t, r, tt.count)
		})
	}
}

func TestListPayloads_Error(t *testing.T) {

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsListParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			t.Context(),
			actions.PayloadsListParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.PayloadsList(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestUpdatePayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		p    actions.PayloadsUpdateParams
	}{
		{
			"update name",
			actions.PayloadsUpdateParams{
				Name:    "payload1",
				NewName: "payload1_updated",
			},
		},
		{
			"update notify protocols",
			actions.PayloadsUpdateParams{
				Name:            "payload1",
				NotifyProtocols: []string{database.ProtoCategoryHTTP},
			},
		},
		{
			"update stored events",
			actions.PayloadsUpdateParams{
				Name:        "payload1",
				StoreEvents: pointer.Bool(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.PayloadsUpdate(ctx, tt.p)
			assert.NoError(t, err)
			assert.NotNil(t, r)

			if tt.p.NewName != "" {
				assert.Equal(t, tt.p.NewName, r.Name)
			}

			if tt.p.NotifyProtocols != nil {
				assert.Equal(t, tt.p.NotifyProtocols, []string(r.NotifyProtocols))
			}

			if tt.p.StoreEvents != nil {
				assert.Equal(t, *tt.p.StoreEvents, r.StoreEvents)
			}
		})
	}
}

func TestUpdatePayload_Error(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsUpdateParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			t.Context(),
			actions.PayloadsUpdateParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty name",
			ctx,
			actions.PayloadsUpdateParams{
				Name: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			ctx,
			actions.PayloadsUpdateParams{
				Name: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.PayloadsUpdate(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestClearPayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name         string
		p            actions.PayloadsClearParams
		deletedCount int
	}{
		{
			"delete all",
			actions.PayloadsClearParams{
				Name: "",
			},
			6,
		},
		{
			"delete some",
			actions.PayloadsClearParams{
				Name: "1",
			},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.PayloadsClear(ctx, tt.p)
			assert.NoError(t, err)
			assert.NotNil(t, r)

			assert.Len(t, r, tt.deletedCount)
		})
	}
}

func TestClearPayloads_Error(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsClearParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			t.Context(),
			actions.PayloadsClearParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.PayloadsClear(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
