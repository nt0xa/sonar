package cmd_test

import (
	"context"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
)

func TestDNSRecordCreate_Success(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		result  actions.DNSRecordsCreateParams
	}{
		{
			"defaults",
			"dns new -p payload -n name 192.168.1.1",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload",
				Name:        "name",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"192.168.1.1"},
				Strategy:    models.DNSStrategyAll,
			},
		},
		{
			"custom 1",
			`dns new -p payload -n name -t mx -l 120 -s round-robin "10 mx.example.com."`,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload",
				Name:        "name",
				TTL:         120,
				Type:        models.DNSTypeMX,
				Values:      []string{"10 mx.example.com."},
				Strategy:    models.DNSStrategyRoundRobin,
			},
		},
		{
			"custom 2",
			`dns new -p payload -n name -t a -l 100 -s rebind 1.1.1.1 2.2.2.2 3.3.3.3`,
			actions.DNSRecordsCreateParams{
				PayloadName: "payload",
				Name:        "name",
				TTL:         100,
				Type:        models.DNSTypeA,
				Values:      []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"},
				Strategy:    models.DNSStrategyRebind,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, acts, hnd := prepare()

			res := &actions.CreateDNSRecordResultData{}

			acts.
				On("DNSRecordsCreate", ctx, tt.result).
				Return(res, nil)

			hnd.On("Handle", ctx, res)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			_, err = c.Exec(ctx, &actions.User{}, args)

			assert.NoError(t, err)

			acts.AssertExpectations(t)
			hnd.AssertExpectations(t)
		})
	}

}

func TestDNSRecordCreate_Error(t *testing.T) {
	c, _, _ := prepare()

	args, err := shlex.Split("dns new -p payload -n name")
	require.NoError(t, err)

	out, err := c.Exec(context.Background(), &actions.User{}, args)
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestDNSRecordDelete_Success(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		result  actions.DNSRecordsDeleteParams
	}{
		{
			"1",
			"dns del -p payload -n name -t a",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload",
				Name:        "name",
				Type:        models.DNSTypeA,
			},
		},
		{
			"2",
			"dns del -p payload -n name -t mx",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload",
				Name:        "name",
				Type:        models.DNSTypeMX,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, acts, hnd := prepare()

			res := actions.DNSRecordsDeleteResult(&actions.DNSRecord{})

			acts.
				On("DNSRecordsDelete", ctx, tt.result).
				Return(res, nil)

			hnd.On("Handle", ctx, res)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			_, err = c.Exec(ctx, &actions.User{}, args)

			assert.NoError(t, err)

			acts.AssertExpectations(t)
			hnd.AssertExpectations(t)
		})
	}
}

func TestDNSList_Success(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		result  actions.DNSRecordsListParams
	}{
		{
			"1",
			"dns list -p payload",
			actions.DNSRecordsListParams{
				PayloadName: "payload",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, acts, hnd := prepare()

			res := &actions.ListDNSRecordsResultData{}

			acts.
				On("DNSRecordsList", ctx, tt.result).
				Return(res, nil)

			hnd.On("Handle", ctx, res)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			_, err = c.Exec(ctx, &actions.User{}, args)

			assert.NoError(t, err)

			acts.AssertExpectations(t)
			hnd.AssertExpectations(t)
		})
	}
}
