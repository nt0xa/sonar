package actionsdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/errors"
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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{"127.0.0.1"},
				Strategy:    string(database.DNSStrategyAll),
			},
		},
		{
			"aaaa",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        string(database.DNSRecordTypeAAAA),
				Values:      []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
				Strategy:    string(database.DNSStrategyAll),
			},
		},
		{
			"mx",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        string(database.DNSRecordTypeMX),
				Values:      []string{"10 mx.example.com."},
				Strategy:    string(database.DNSStrategyAll),
			},
		},
		{
			"txt",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        string(database.DNSRecordTypeTXT),
				Values:      []string{"test string"},
				Strategy:    string(database.DNSStrategyAll),
			},
		},
		{
			"cname",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         60,
				Type:        string(database.DNSRecordTypeTXT),
				Values:      []string{"test.example.com."},
				Strategy:    string(database.DNSStrategyAll),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			u, err := db.UsersGetByID(t.Context(), 1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(t.Context(), u)

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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{"127.0.0.1"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{"127.0.0.1"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{"127.0.0.1"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{"invalid"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeAAAA),
				Values:      []string{"invalid"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeMX),
				Values:      []string{"10 example.com"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeCNAME),
				Values:      []string{"example.com"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{"127.0.0.1"},
				Strategy:    string(database.DNSStrategyAll),
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
				Type:        string(database.DNSRecordTypeA),
				Values:      []string{"127.0.0.1"},
				Strategy:    string(database.DNSStrategyAll),
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := t.Context()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(t.Context(), 1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(t.Context(), u)
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

			u, err := db.UsersGetByID(t.Context(), 1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(t.Context(), u)

			_, err = acts.DNSRecordsDelete(ctx, tt.p)
			assert.NoError(t, err)

			p, err := db.PayloadsGetByUserAndName(t.Context(), u.ID, tt.p.PayloadName)
			assert.NoError(t, err)

			_, err = db.DNSRecordsGetByPayloadNameAndType(t.Context(), database.DNSRecordsGetByPayloadNameAndTypeParams{
				PayloadID: p.ID,
				Name:      tt.name,
				Type:      database.DNSRecordType(tt.typ),
			})
			assert.Error(t, err, database.ErrNoRows)
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

			ctx := t.Context()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(t.Context(), 1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(t.Context(), u)
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

			u, err := db.UsersGetByID(t.Context(), 1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(t.Context(), u)

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

			ctx := t.Context()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(t.Context(), 1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(t.Context(), u)
			}

			_, err := acts.DNSRecordsList(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDNSRecordsClear_Success(t *testing.T) {

	tests := []struct {
		name  string
		p     actions.DNSRecordsClearParams
		count int
	}{
		{
			"payload1",
			actions.DNSRecordsClearParams{
				PayloadName: "payload1",
			},
			9,
		},
		{
			"payload1",
			actions.DNSRecordsClearParams{
				PayloadName: "payload1",
				Name:        "test-a",
			},
			1,
		},
		{
			"payload4",
			actions.DNSRecordsClearParams{
				PayloadName: "payload4",
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			u, err := db.UsersGetByID(t.Context(), 1)
			require.NoError(t, err)

			ctx := actionsdb.SetUser(t.Context(), u)

			list, err := acts.DNSRecordsClear(ctx, tt.p)
			assert.NoError(t, err)
			assert.Len(t, list, tt.count)
		})
	}
}

func TestDNSRecordsClear_Error(t *testing.T) {

	tests := []struct {
		name   string
		userID int
		p      actions.DNSRecordsClearParams
		err    error
	}{
		{
			"no user in ctx",
			0,
			actions.DNSRecordsClearParams{
				Name: "test",
			},
			&errors.InternalError{},
		},
		{
			"not existing payload",
			1,
			actions.DNSRecordsClearParams{
				PayloadName: "not-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := t.Context()
			if tt.userID != 0 {
				u, err := db.UsersGetByID(t.Context(), 1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(t.Context(), u)
			}

			_, err := acts.DNSRecordsClear(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
