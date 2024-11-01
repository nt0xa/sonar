package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	actions_mock "github.com/nt0xa/sonar/internal/actions/mock"
	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/pointer"
)

var (
	ctx = context.Background()
)

type ResultMock struct {
	mock.Mock
}

func (m *ResultMock) OnResult(res actions.Result) error {
	m.Called(res)
	return nil
}

func prepare() (*cmd.Command, *actions_mock.Actions, *ResultMock) {
	actions := &actions_mock.Actions{}
	res := &ResultMock{}

	c := cmd.New(actions)

	return c, actions, res
}

func ptr[T any](v T) *T {
	return &v
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
			&actions.PayloadsCreateResult{},
		},

		// List

		{
			"list substr",
			"PayloadsList",
			actions.PayloadsListParams{
				Name: "substr",
			},
			actions.PayloadsListResult{},
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
			&actions.PayloadsUpdateResult{},
		},

		// Delete

		{
			"del test",
			"PayloadsDelete",
			actions.PayloadsDeleteParams{
				Name: "test",
			},
			&actions.PayloadsDeleteResult{},
		},

		// Clear

		{
			"clr",
			"PayloadsClear",
			actions.PayloadsClearParams{
				Name: "",
			},
			actions.PayloadsClearResult{},
		},
		{
			"clr test",
			"PayloadsClear",
			actions.PayloadsClearParams{
				Name: "test",
			},
			actions.PayloadsClearResult{},
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
			&actions.DNSRecordsCreateResult{},
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
			&actions.DNSRecordsCreateResult{},
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
			&actions.DNSRecordsCreateResult{},
		},

		// List

		{
			"dns list -p payload",
			"DNSRecordsList",
			actions.DNSRecordsListParams{
				PayloadName: "payload",
			},
			actions.DNSRecordsListResult{},
		},

		// Delete

		{
			"dns del -p payload 1",
			"DNSRecordsDelete",
			actions.DNSRecordsDeleteParams{
				PayloadName: "payload",
				Index:       1,
			},
			&actions.DNSRecordsDeleteResult{},
		},

		// Clear

		{
			"dns clr -p payload1 -n test",
			"DNSRecordsClear",
			actions.DNSRecordsClearParams{
				PayloadName: "payload1",
				Name:        "test",
			},
			actions.DNSRecordsClearResult{},
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
			&actions.UsersCreateResult{},
		},

		// Delete

		{
			"users del test",
			"UsersDelete",
			actions.UsersDeleteParams{
				Name: "test",
			},
			&actions.UsersDeleteResult{},
		},

		//
		// User
		//

		{
			"profile",
			"ProfileGet",
			nil,
			&actions.ProfileGetResult{},
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
			actions.EventsListResult{},
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
			actions.EventsListResult{},
		},

		// Get

		{
			"events get -p test 5",
			"EventsGet",
			actions.EventsGetParams{
				PayloadName: "test",
				Index:       5,
			},
			&actions.EventsGetResult{},
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
			&actions.HTTPRoutesCreateResult{},
		},

		// Update

		{
			"http -p payload mod 1 -m POST -P /test -c 201 -H 'Content-Type: application/json' -d -b test",
			"HTTPRoutesUpdate",
			actions.HTTPRoutesUpdateParams{
				Payload: "payload",
				Index:   1,
				Method:  ptr("POST"),
				Path:    ptr("/test"),
				Code:    ptr(201),
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body:      ptr("dGVzdA=="),
				IsDynamic: ptr(true),
			},
			&actions.HTTPRoutesUpdateResult{},
		},

		// List

		{
			"http list -p payload",
			"HTTPRoutesList",
			actions.HTTPRoutesListParams{
				PayloadName: "payload",
			},
			actions.HTTPRoutesListResult{},
		},

		// Delete

		{
			"http del -p payload 1",
			"HTTPRoutesDelete",
			actions.HTTPRoutesDeleteParams{
				PayloadName: "payload",
				Index:       1,
			},
			&actions.HTTPRoutesDeleteResult{},
		},

		// Clear

		{
			"http clr -p payload1 -P /test",
			"HTTPRoutesClear",
			actions.HTTPRoutesClearParams{
				PayloadName: "payload1",
				Path:        "/test",
			},
			actions.HTTPRoutesClearResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			c, acts, res := prepare()

			if tt.params != nil {
				acts.
					On(tt.action, ctx, tt.params).
					Return(tt.result, nil)
			} else {
				acts.
					On(tt.action, ctx).
					Return(tt.result, nil)
			}

			acts.On("ProfileGet", ctx).
				Return(&actions.ProfileGetResult{User: actions.User{IsAdmin: true}}, nil)

			res.On("OnResult", tt.result)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			c.Exec(ctx, args, res.OnResult)

			acts.AssertExpectations(t)
			res.AssertExpectations(t)
		})
	}

}
