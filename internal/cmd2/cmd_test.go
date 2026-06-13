package cmd2_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/cmd2"
	"github.com/nt0xa/sonar/internal/service"
	service_mock "github.com/nt0xa/sonar/internal/service/mock"
)

func ptr[T any](v T) *T { return &v }

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

func TestCmd(t *testing.T) {
	tests := []struct {
		cmdline string
		method  string
		input   any
		result  any
	}{
		// Payloads
		{
			"new test -p dns,http -e",
			"PayloadsCreate",
			service.PayloadsCreateInput{
				Name:            "test",
				NotifyProtocols: []service.ProtoCategory{service.ProtoCategoryDns, service.ProtoCategoryHttp},
				StoreEvents:     true,
			},
			&service.Payload{Name: "test"},
		},
		{
			"list foo -p 2 -s 20",
			"PayloadsList",
			service.PayloadsListInput{Name: "foo", Page: 2, PerPage: 20},
			[]service.Payload{{Name: "foo"}},
		},
		{
			"mod old -n new -e",
			"PayloadsUpdate",
			service.PayloadsUpdateInput{Name: "old", NewName: "new", StoreEvents: ptr(true)},
			&service.Payload{Name: "new"},
		},
		{
			"del foo",
			"PayloadsDelete",
			service.PayloadsDeleteInput{Name: "foo"},
			&service.Payload{Name: "foo"},
		},
		{
			"clr foo",
			"PayloadsClear",
			service.PayloadsClearInput{Name: "foo"},
			[]service.Payload{{Name: "foo"}},
		},

		// DNS records
		{
			"dns new 1.1.1.1 -p test -n www",
			"DNSRecordsCreate",
			service.DNSRecordsCreateInput{
				PayloadName: "test",
				Name:        "www",
				TTL:         60,
				Type:        service.DNSRecordTypeA,
				Values:      []string{"1.1.1.1"},
				Strategy:    service.DNSRecordStrategyAll,
			},
			&service.DNSRecord{Index: 1},
		},
		{
			"dns del 1 -p test",
			"DNSRecordsDelete",
			service.DNSRecordsDeleteInput{PayloadName: "test", Index: 1},
			&service.DNSRecord{Index: 1},
		},
		{
			"dns list -p test",
			"DNSRecordsList",
			service.DNSRecordsListInput{PayloadName: "test"},
			[]service.DNSRecord{{Index: 1}},
		},
		{
			"dns clr -p test -n www",
			"DNSRecordsClear",
			service.DNSRecordsClearInput{PayloadName: "test", Name: "www"},
			[]service.DNSRecord{{Index: 1}},
		},

		// HTTP routes
		{
			"http new body -p test -m POST -P /x -c 201",
			"HTTPRoutesCreate",
			service.HTTPRoutesCreateInput{
				PayloadName: "test",
				Method:      service.HTTPMethodPOST,
				Path:        "/x",
				Code:        201,
				Headers:     map[string][]string{},
				Body:        b64("body"),
				IsDynamic:   false,
			},
			&service.HTTPRoute{Index: 1},
		},
		{
			"http mod 1 -p test -c 404",
			"HTTPRoutesUpdate",
			service.HTTPRoutesUpdateInput{Payload: "test", Index: 1, Code: ptr(404)},
			&service.HTTPRoute{Index: 1},
		},
		{
			"http del 1 -p test",
			"HTTPRoutesDelete",
			service.HTTPRoutesDeleteInput{PayloadName: "test", Index: 1},
			&service.HTTPRoute{Index: 1},
		},
		{
			"http list -p test",
			"HTTPRoutesList",
			service.HTTPRoutesListInput{PayloadName: "test"},
			[]service.HTTPRoute{{Index: 1}},
		},
		{
			"http clr -p test -P /x",
			"HTTPRoutesClear",
			service.HTTPRoutesClearInput{PayloadName: "test", Path: "/x"},
			[]service.HTTPRoute{{Index: 1}},
		},

		// Events
		{
			"events list -p test -l 5 -o 2",
			"EventsList",
			service.EventsListInput{PayloadName: "test", Limit: 5, Offset: 2},
			[]service.Event{{Index: 1}},
		},
		{
			"events get 3 -p test",
			"EventsGet",
			service.EventsGetInput{PayloadName: "test", Index: 3},
			&service.Event{Index: 3},
		},

		// Users (admin)
		{
			"users new bob -a --token tok",
			"UsersCreate",
			service.UsersCreateInput{Name: "bob", IsAdmin: true, APIToken: ptr("tok")},
			&service.User{Name: "bob"},
		},
		{
			"users del bob",
			"UsersDelete",
			service.UsersDeleteInput{Name: "bob"},
			&service.User{Name: "bob"},
		},

		// Audit (admin)
		{
			"audit list --action create",
			"AuditRecordsList",
			service.AuditRecordsListInput{Action: service.AuditActionCreate, Page: 1, PerPage: 50},
			[]service.AuditRecord{{ID: 1}},
		},
		{
			"audit get 7",
			"AuditRecordsGet",
			service.AuditRecordsGetInput{ID: 7},
			&service.AuditRecord{ID: 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.cmdline, func(t *testing.T) {
			svc := &service_mock.ServerService{}
			// Auth check runs ProfileGet before every command; admin commands
			// require IsAdmin.
			svc.On("ProfileGet", mock.Anything).
				Return(&service.User{IsAdmin: true}, nil)
			svc.On(tt.method, mock.Anything, tt.input).
				Return(tt.result, nil)

			args, err := shlex.Split(tt.cmdline)
			require.NoError(t, err)

			res, err := cmd2.New(svc).Exec(context.Background(), args)
			require.NoError(t, err)
			require.Equal(t, tt.result, res)

			svc.AssertCalled(t, tt.method, mock.Anything, tt.input)
		})
	}
}

func TestProfileGet(t *testing.T) {
	svc := &service_mock.ServerService{}
	user := &service.User{Name: "me", IsAdmin: true}
	svc.On("ProfileGet", mock.Anything).Return(user, nil)

	res, err := cmd2.New(svc).Exec(context.Background(), []string{"profile"})
	require.NoError(t, err)
	require.Equal(t, user, res)
}

func TestAdminCheck(t *testing.T) {
	svc := &service_mock.ServerService{}
	// Non-admin profile: admin-only commands must be rejected before the service
	// method is called.
	svc.On("ProfileGet", mock.Anything).Return(&service.User{IsAdmin: false}, nil)

	_, err := cmd2.New(svc).Exec(context.Background(), []string{"users", "del", "bob"})
	require.Error(t, err)

	var svcErr service.Error
	require.ErrorAs(t, err, &svcErr)
	require.Equal(t, service.ErrorKindForbidden, svcErr.Kind)

	svc.AssertNotCalled(t, "UsersDelete", mock.Anything, mock.Anything)
}
