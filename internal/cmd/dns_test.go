package cmd_test

import (
	"context"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
)

func TestDNSRecordCreate_Success(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		result  actions.CreateDNSRecordParams
	}{
		{
			"defaults",
			"dns new -p payload -n name 192.168.1.1",
			actions.CreateDNSRecordParams{
				"payload", "name", 60, models.DNSTypeA, []string{"192.168.1.1"}, models.DNSStrategyAll,
			},
		},
		{
			"custom 1",
			`dns new -p payload -n name -t mx -l 120 -s round-robin "10 mx.example.com."`,
			actions.CreateDNSRecordParams{
				"payload", "name", 120, models.DNSTypeMX, []string{"10 mx.example.com."}, models.DNSStrategyRoundRobin,
			},
		},
		{
			"custom 2",
			`dns new -p payload -n name -t a -l 100 -s rebind 1.1.1.1 2.2.2.2 3.3.3.3`,
			actions.CreateDNSRecordParams{
				"payload", "name", 100, models.DNSTypeA, []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"}, models.DNSStrategyRebind,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, acts, hnd := prepare()

			ctx := actions.SetUser(context.Background(), user)

			acts.
				On("CreateDNSRecord", ctx, tt.result).
				Return(&actions.CreateDNSRecordResultData{}, nil)

			hnd.On("Handle", mock.Anything, mock.Anything)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			_, err = cmd.Exec(context.Background(), user, args)

			assert.NoError(t, err)
		})
	}
}

func TestDNSRecordCreate_Error(t *testing.T) {
	cmd, _, _ := prepare()

	args, err := shlex.Split("dns new -p payload -n name")
	require.NoError(t, err)

	out, err := cmd.Exec(context.Background(), user, args)
	assert.Error(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "required")
}

func TestDNSDeleteCreate_Success(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		result  actions.DeleteDNSRecordParams
	}{
		{
			"1",
			"dns del -p payload -n name -t a",
			actions.DeleteDNSRecordParams{
				"payload", "name", models.DNSTypeA,
			},
		},
		{
			"2",
			"dns del -p payload -n name -t mx",
			actions.DeleteDNSRecordParams{
				"payload", "name", models.DNSTypeMX,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, acts, hnd := prepare()

			ctx := actions.SetUser(context.Background(), user)

			acts.
				On("DeleteDNSRecord", ctx, tt.result).
				Return(&actions.MessageResult{}, nil)

			hnd.On("Handle", mock.Anything, mock.Anything)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			_, err = cmd.Exec(context.Background(), user, args)

			assert.NoError(t, err)
		})
	}
}

func TestDNSList_Success(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		result  actions.ListDNSRecordsParams
	}{
		{
			"1",
			"dns list -p payload",
			actions.ListDNSRecordsParams{
				"payload",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, acts, hnd := prepare()

			ctx := actions.SetUser(context.Background(), user)

			acts.
				On("ListDNSRecords", ctx, tt.result).
				Return(&actions.ListDNSRecordsResultData{}, nil)

			hnd.On("Handle", mock.Anything, mock.Anything)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			_, err = cmd.Exec(context.Background(), user, args)

			assert.NoError(t, err)
		})
	}
}
