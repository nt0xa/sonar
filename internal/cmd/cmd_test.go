package cmd_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	actions_mock "github.com/bi-zone/sonar/internal/actions/mock"
	"github.com/bi-zone/sonar/internal/cmd"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ctx = context.WithValue(context.Background(), "key", "value")
)

func prepare() (cmd.Command, *actions_mock.Actions, *actions_mock.ResultHandler) {
	actions := &actions_mock.Actions{}
	handler := &actions_mock.ResultHandler{}

	c := cmd.New(actions, handler, nil)

	return c, actions, handler
}

func TestCmd(t *testing.T) {
	tests := []struct {
		comment string
		cmdline string
		action  string
		params  interface{}
		result  interface{}
	}{

		//
		// Payloads
		//

		// Create

		{
			"1",
			"new test -p dns,http",
			"PayloadsCreate",
			actions.PayloadsCreateParams{
				Name:            "test",
				NotifyProtocols: []string{models.PayloadProtocolDNS, models.PayloadProtocolHTTP},
			},
			(actions.PayloadsCreateResult)(nil),
		},

		// List

		{
			"1",
			"list substr",
			"PayloadsList",
			actions.PayloadsListParams{
				Name: "substr",
			},
			(actions.PayloadsListResult)(nil),
		},

		// Update

		{
			"1",
			"mod -n new -p dns old",
			"PayloadsUpdate",
			actions.PayloadsUpdateParams{
				Name:            "old",
				NewName:         "new",
				NotifyProtocols: []string{models.PayloadProtocolDNS},
			},
			(actions.PayloadsUpdateResult)(nil),
		},

		// Delete

		{
			"1",
			"del test",
			"PayloadsDelete",
			actions.PayloadsDeleteParams{
				Name: "test",
			},
			(actions.PayloadsDeleteResult)(nil),
		},

		//
		// DNS
		//

		// Create

		{
			"1",
			"dns new -p payload -n name 192.168.1.1",
			"DNSRecordsCreate",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload",
				Name:        "name",
				TTL:         60,
				Type:        models.DNSTypeA,
				Values:      []string{"192.168.1.1"},
				Strategy:    models.DNSStrategyAll,
			},
			(actions.DNSRecordsCreateResult)(nil),
		},
		{
			"2",
			`dns new -p payload -n name -t mx -l 120 -s round-robin "10 mx.example.com."`,
			"DNSRecordsCreate",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload",
				Name:        "name",
				TTL:         120,
				Type:        strings.ToLower(models.DNSTypeMX),
				Values:      []string{"10 mx.example.com."},
				Strategy:    models.DNSStrategyRoundRobin,
			},
			(actions.DNSRecordsCreateResult)(nil),
		},
		{
			"3",
			`dns new -p payload -n name -t a -l 100 -s rebind 1.1.1.1 2.2.2.2 3.3.3.3`,
			"DNSRecordsCreate",
			actions.DNSRecordsCreateParams{
				PayloadName: "payload",
				Name:        "name",
				TTL:         100,
				Type:        strings.ToLower(models.DNSTypeA),
				Values:      []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"},
				Strategy:    models.DNSStrategyRebind,
			},
			(actions.DNSRecordsCreateResult)(nil),
		},

		// List

		{
			"1",
			"dns list -p payload",
			"DNSRecordsList",
			actions.DNSRecordsListParams{
				PayloadName: "payload",
			},
			(actions.DNSRecordsListResult)(nil),
		},

		// Delete

		{
			"1",
			"dns del -p payload -n name -t a",
			"DNSRecordsDelete",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload",
				Name:        "name",
				Type:        strings.ToLower(models.DNSTypeA),
			},
			(actions.DNSRecordsDeleteResult)(nil),
		},

		//
		// Users
		//

		// Create

		{
			"1",
			"users new -a -p telegram.id=1337 -p api.token=token test",
			"UsersCreate",
			actions.UsersCreateParams{
				Name: "test",
				Params: models.UserParams{
					TelegramID: 1337,
					APIToken:   "token",
				},
				IsAdmin: true,
			},
			(actions.UsersCreateResult)(nil),
		},

		// Delete

		{
			"1",
			"users del test",
			"UsersDelete",
			actions.UsersDeleteParams{
				Name: "test",
			},
			(actions.UsersDeleteResult)(nil),
		},

		//
		// User
		//

		{
			"1",
			"user",
			"UserCurrent",
			nil,
			(actions.UserCurrentResult)(nil),
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.action, tt.comment), func(t *testing.T) {
			c, acts, hnd := prepare()

			if tt.params != nil {
				acts.
					On(tt.action, ctx, tt.params).
					Return(tt.result, nil)
			} else {
				acts.
					On(tt.action, ctx).
					Return(tt.result, nil)
			}

			hnd.On(tt.action, ctx, tt.result)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			_, err = c.Exec(ctx, &actions.User{IsAdmin: true}, args)

			assert.NoError(t, err)

			acts.AssertExpectations(t)
			hnd.AssertExpectations(t)
		})
	}

}
