package actions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func TestCreateDNSRecord_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	tests := []struct {
		name string
		p    actions.CreateDNSRecordParams
	}{
		{
			"a",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeA, []string{"127.0.0.1"}, models.DNSStrategyAll,
			},
		},
		{
			"aaaa",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeAAAA, []string{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}, models.DNSStrategyAll,
			},
		},
		{
			"mx",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeMX, []string{"10 mx.example.com."}, models.DNSStrategyAll,
			},
		},
		{
			"txt",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeTXT, []string{"test string"}, models.DNSStrategyAll,
			},
		},
		{
			"cname",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeCNAME, []string{"test.example.com."}, models.DNSStrategyAll,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.CreateDNSRecord(u, tt.p)
			require.NoError(t, err)
			require.NotNil(t, r)

			assert.Equal(t, tt.p.Name, r.Record.Name)
		})
	}
}

func TestCreateDNSRecord_Error(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	tests := []struct {
		name string
		p    actions.CreateDNSRecordParams
		err  errors.Error
	}{
		{
			"empty name",
			actions.CreateDNSRecordParams{
				"payload1", "", 60, models.DNSTypeA, []string{"127.0.0.1"}, models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"empty payload name",
			actions.CreateDNSRecordParams{
				"", "test", 60, models.DNSTypeA, []string{"127.0.0.1"}, models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"empty values",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeA, []string{}, models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid a",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeA, []string{"invalid"}, models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid aaaa",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeAAAA, []string{"invalid"}, models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid mx",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeMX, []string{"10 example.com"}, models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"invalid cname",
			actions.CreateDNSRecordParams{
				"payload1", "test", 60, models.DNSTypeMX, []string{"example.com"}, models.DNSStrategyAll,
			},
			&errors.ValidationError{},
		},
		{
			"not existing payload name",
			actions.CreateDNSRecordParams{
				"not-exists", "test", 60, models.DNSTypeA, []string{"127.0.0.1"}, models.DNSStrategyAll,
			},
			&errors.NotFoundError{},
		},
		{
			"duplicate name and type",
			actions.CreateDNSRecordParams{
				"payload1", "test-a", 60, models.DNSTypeA, []string{"127.0.0.1"}, models.DNSStrategyAll,
			},
			&errors.ConflictError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.CreateDNSRecord(u, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestDeleteDNSRecord_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	tests := []struct {
		name string
		p    actions.DeleteDNSRecordParams
	}{
		{
			"test-a",
			actions.DeleteDNSRecordParams{
				"payload1", "test-a", models.DNSTypeA,
			},
		},
		{
			"test-aaaa",
			actions.DeleteDNSRecordParams{
				"payload1", "test-aaaa", models.DNSTypeAAAA,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DeleteDNSRecord(u, tt.p)
			assert.NoError(t, err)
		})
	}
}

func TestDeleteDNSRecord_Error(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	tests := []struct {
		name string
		p    actions.DeleteDNSRecordParams
		err  errors.Error
	}{
		{
			"empty payload name",
			actions.DeleteDNSRecordParams{"", "test-a", models.DNSTypeA},
			&errors.ValidationError{},
		},
		{
			"empty name",
			actions.DeleteDNSRecordParams{"test", "", models.DNSTypeA},
			&errors.ValidationError{},
		},
		{
			"invalid record type",
			actions.DeleteDNSRecordParams{"payload1", "test-a", "invalid"},
			&errors.ValidationError{},
		},
		{
			"not existing payload",
			actions.DeleteDNSRecordParams{"not-exists", "test-a", models.DNSTypeA},
			&errors.NotFoundError{},
		},
		{
			"not existing record",
			actions.DeleteDNSRecordParams{"payload1", "not-exists", models.DNSTypeA},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			_, err := acts.DeleteDNSRecord(u, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestListDNSRecords_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	p := actions.ListDNSRecordsParams{
		PayloadName: "payload1",
	}

	r, err := acts.ListDNSRecords(u, p)
	require.NoError(t, err)
	require.NotNil(t, r)

	assert.Len(t, r.Records, 9)
}
