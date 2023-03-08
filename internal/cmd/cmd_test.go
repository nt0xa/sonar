package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/russtone/sonar/internal/actions"
	actions_mock "github.com/russtone/sonar/internal/actions/mock"
	"github.com/russtone/sonar/internal/cmd"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/utils/pointer"
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
			"new test -p dns,http -e",
			"PayloadsCreate",
			actions.PayloadsCreateParams{
				Name: "test",
				NotifyProtocols: []string{
					models.ProtoCategoryDNS.String(),
					models.ProtoCategoryHTTP.String(),
				},
				StoreEvents: true,
			},
			(actions.PayloadsCreateResult)(nil),
		},

		// List

		{
			"list substr",
			"PayloadsList",
			actions.PayloadsListParams{
				Name: "substr",
			},
			(actions.PayloadsListResult)(nil),
		},

		// Update

		{
			"mod -n new -p dns old -e=false",
			"PayloadsUpdate",
			actions.PayloadsUpdateParams{
				Name:            "old",
				NewName:         "new",
				NotifyProtocols: []string{models.ProtoCategoryDNS.String()},
				StoreEvents:     pointer.Bool(false),
			},
			(actions.PayloadsUpdateResult)(nil),
		},

		// Delete

		{
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
			"dns list -p payload",
			"DNSRecordsList",
			actions.DNSRecordsListParams{
				PayloadName: "payload",
			},
			(actions.DNSRecordsListResult)(nil),
		},

		// Delete

		{
			"dns del -p payload 1",
			"DNSRecordsDelete",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload",
				Index:       1,
			},
			(actions.DNSRecordsDeleteResult)(nil),
		},

		//
		// Users
		//

		// Create

		{
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
			"user",
			"UserCurrent",
			nil,
			(actions.UserCurrentResult)(nil),
		},

		//
		// Events
		//

		// List

		{
			"events list -p test -c 5 -a 3",
			"EventsList",
			actions.EventsListParams{
				PayloadName: "test",
				Count:       5,
				After:       3,
			},
			(actions.EventsListResult)(nil),
		},
		{
			"events list -p test -c 5 -b 3 -r",
			"EventsList",
			actions.EventsListParams{
				PayloadName: "test",
				Count:       5,
				Before:      3,
				Reverse:     true,
			},
			(actions.EventsListResult)(nil),
		},

		// Get

		{
			"events get -p test 5",
			"EventsGet",
			actions.EventsGetParams{
				PayloadName: "test",
				Index:       5,
			},
			(actions.EventsGetResult)(nil),
		},

		//
		// HTTP
		//

		// Create

		{
			"http -p payload new -m POST -P /test -c 201 -H 'Content-Type: application/json' -d test",
			"HTTPRoutesCreate",
			actions.HTTPRoutesCreateParams{
				PayloadName: "payload",
				Method:      "POST",
				Path:        "/test",
				Code:        201,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body:      "dGVzdA==",
				IsDynamic: true,
			},
			(actions.HTTPRoutesCreateResult)(nil),
		},

		// List

		{
			"http list -p payload",
			"HTTPRoutesList",
			actions.HTTPRoutesListParams{
				PayloadName: "payload",
			},
			(actions.HTTPRoutesListResult)(nil),
		},

		// Delete

		{
			"http del -p payload 1",
			"HTTPRoutesDelete",
			actions.HTTPRoutesDeleteParams{
				PayloadName: "payload",
				Index:       1,
			},
			(actions.HTTPRoutesDeleteResult)(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
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

			_, err = c.Exec(ctx, &actions.User{IsAdmin: true}, true, args)

			assert.NoError(t, err)

			acts.AssertExpectations(t)
			hnd.AssertExpectations(t)
		})
	}

}
