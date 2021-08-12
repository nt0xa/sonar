package actionsdb_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/actionsdb"
	"github.com/bi-zone/sonar/internal/database/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func TestDNSRecordsCreate_Success(t *testing.T) {

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

			u, err := db.UsersGetByID(1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(context.Background(), u)

			r, err := acts.DNSRecordsCreate(ctx, tt.p)
			require.NoError(t, err)
			require.NotNil(t, r)

			assert.Equal(t, tt.p.Name, r.Name)
		})
	}
}

func TestDNSRecordsCreate_Error(t *testing.T) {

	tests := []struct {
		name   string
		userID int
		p      actions.DNSRecordsCreateParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
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
			1,
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
			1,
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
			1,
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
			1,
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
			1,
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
			1,
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
			1,
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
			1,
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
			1,
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

			ctx := context.Background()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.DNSRecordsCreate(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDNSRecordsDelete_Success(t *testing.T) {

	tests := []struct {
		name string
		typ  string
		p    actions.DNSRecordsDeleteParams
	}{
		{
			"test-a",
			"A",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Index:       1,
			},
		},
		{
			"test-aaaa",
			"AAAA",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Index:       2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			u, err := db.UsersGetByID(1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(context.Background(), u)

			_, err = acts.DNSRecordsDelete(ctx, tt.p)
			assert.NoError(t, err)

			p, err := db.PayloadsGetByUserAndName(u.ID, tt.p.PayloadName)
			assert.NoError(t, err)

			_, err = db.DNSRecordsGetByPayloadNameAndType(p.ID, tt.name, tt.typ)
			assert.Error(t, err, sql.ErrNoRows)
		})
	}
}

func TestDNSRecordsDelete_Error(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		p      actions.DNSRecordsDeleteParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Index:       1,
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			1,
			actions.DNSRecordsDeleteParams{
				PayloadName: "",
				Index:       1,
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			1,
			actions.DNSRecordsDeleteParams{
				PayloadName: "not-exist",
				Index:       1,
			},
			&errors.NotFoundError{},
		},
		{
			"not existing index",
			1,
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload1",
				Index:       1337,
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := context.Background()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.DNSRecordsDelete(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDNSRecordsList_Success(t *testing.T) {

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

			u, err := db.UsersGetByID(1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(context.Background(), u)

			list, err := acts.DNSRecordsList(ctx, tt.p)
			assert.NoError(t, err)
			assert.Len(t, list, tt.count)
		})
	}
}

func TestDNSRecordsList_Error(t *testing.T) {

	tests := []struct {
		name   string
		userID int
		p      actions.DNSRecordsListParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
			actions.DNSRecordsListParams{
				PayloadName: "payload1",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			1,
			actions.DNSRecordsListParams{
				PayloadName: "",
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			1,
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

			ctx := context.Background()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.DNSRecordsList(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
