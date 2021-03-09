package actionsdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/actionsdb"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func TestCreateDNSRecord_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		p    actions.DNSRecordsCreateParams
	}{
		{
			"a",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    models.DNSStrategyAll,
			},
		},
		{
			"aaaa",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeAAAA,
				Values:      []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				Strategy:    models.DNSStrategyAll,
			},
		},
		{
			"mx",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeMX,
				Values:      []string{"10 mx.example.com."},
				Strategy:    models.DNSStrategyAll,
			},
		},
		{
			"txt",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeTXT,
				Values:      []string{"test string"},
				Strategy:    models.DNSStrategyAll,
			},
		},
		{
			"cname",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeTXT,
				Values:      []string{"test.example.com."},
				Strategy:    models.DNSStrategyAll,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.DNSRecordsCreate(ctx, tt.p)
			require.NoError(t, err)
			require.NotNil(t, r)

			assert.Equal(t, tt.p.Name, r.Record.Name)
		})
	}
}

func TestCreateDNSRecord_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.DNSRecordsCreateParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.InternalError{},
		},
		{
			"empty name",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"empty payload name",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"empty values",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid a",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"invalid"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid aaaa",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeAAAA,
				Values:      []string{"invalid"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid mx",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeMX,
				Values:      []string{"10 example.com"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid cname",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeCNAME,
				Values:      []string{"example.com"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload name",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "not-exist",
				Name:        "test",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.NotFoundError{},
		},
		{
			"duplicate name and type",
			ctx,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test-a",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    models.DNSStrategyAll,
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DNSRecordsCreate(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDeleteDNSRecord_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		p    actions.DNSRecordsDeleteParams
	}{
		{
			"test-a",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Name:        "test-a",
				Type:        models.DNSTypeA,
			},
		},
		{
			"test-aaaa",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Name:        "test-aaaa",
				Type:        models.DNSTypeAAAA,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DNSRecordsDelete(ctx, tt.p)
			assert.NoError(t, err)
		})
	}
}

func TestDeleteDNSRecord_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.DNSRecordsDeleteParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Name:        "test-a",
				Type:        models.DNSTypeA,
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			ctx,
			actions.DNSRecordsDeleteParams{
				PayloadName: "",
				Name:        "test-a",
				Type:        models.DNSTypeA,
			},
			&errors.ValidationError{},
		},
		{
			"empty name",
			ctx,
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Name:        "",
				Type:        models.DNSTypeA,
			},
			&errors.ValidationError{},
		},
		{
			"invalid record type",
			ctx,
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Name:        "test-a",
				Type:        "invalid",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			ctx,
			actions.DNSRecordsDeleteParams{
				PayloadName: "not-exist",
				Name:        "test-a",
				Type:        models.DNSTypeA,
			},
			&errors.NotFoundError{},
		},
		{
			"not existing record",
			ctx,
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Name:        "not-exist",
				Type:        models.DNSTypeA,
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DNSRecordsDelete(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestListDNSRecords_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name  string
		p     actions.DNSRecordsListParams
		count int
	}{
		{
			"payload1",
			actions.DNSRecordsListParams{
				PayloadName: "payload1",
			},
			9,
		},
		{
			"payload4",
			actions.DNSRecordsListParams{
				PayloadName: "payload4",
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DNSRecordsList(ctx, tt.p)
			assert.NoError(t, err)
		})
	}
}

func TestListDNSRecords_Error(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

	tests := []struct {
		name string
		ctx  context.Context
		p    actions.DNSRecordsListParams
		err  errors.Error
	}{
		{
			"no user in ctx",
			context.Background(),
			actions.DNSRecordsListParams{
				PayloadName: "payload1",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			ctx,
			actions.DNSRecordsListParams{
				PayloadName: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			ctx,
			actions.DNSRecordsListParams{
				PayloadName: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DNSRecordsList(tt.ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
