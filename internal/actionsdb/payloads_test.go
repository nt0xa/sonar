package actionsdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/actionsdb"
	"github.com/bi-zone/sonar/internal/database/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func TestCreatePayload_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

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
				NotifyProtocols: []string{models.ProtoCategoryDNS.String()},
			},
		},
		{
			"store events",
			actions.PayloadsCreateParams{
				Name:            "test-dns",
				StoreEvents:     100,
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
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsCreateParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
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
		{
			"invalid store events",
			ctx,
			actions.PayloadsCreateParams{
				Name:        "test",
				StoreEvents: 999999,
			},
			&errors.ValidationError{},
		},
		{
			"invalid store events",
			ctx,
			actions.PayloadsCreateParams{
				Name:        "test",
				StoreEvents: -1,
			},
			&errors.ValidationError{},
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
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

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
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsDeleteParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
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
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

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
			2,
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
			context.Background(),
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
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

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
				NotifyProtocols: []string{models.ProtoCategoryHTTP.String()},
			},
		},
		{
			"update stored events",
			actions.PayloadsUpdateParams{
				Name:        "payload1",
				StoreEvents: 8,
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

			if tt.p.StoreEvents >= 0 {
				assert.Equal(t, tt.p.StoreEvents, r.StoreEvents)
			}
		})
	}
}

func TestUpdatePayload_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.PayloadsUpdateParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
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
			"invalid stored events",
			ctx,
			actions.PayloadsUpdateParams{
				Name:        "payload1",
				StoreEvents: -1,
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
