package remotesvc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/modules/api"
	"github.com/nt0xa/sonar/internal/service"
	service_mock "github.com/nt0xa/sonar/internal/service/mock"
	"github.com/nt0xa/sonar/internal/service/remotesvc"
)

// These tests exercise the full client->server->mock round-trip: the real
// remotesvc client talks over HTTP to the real api.Handler(), whose backend is
// a mocked service.ServerService. They verify that client and server agree on
// the wire contract (paths, queries, bodies, status codes, error shape) without
// touching a database.

const (
	AdminToken = "admin-token"
	User1Token = "user1-token"
)

// mustTime parses an RFC3339 timestamp the way the client formats and the
// server parses, so an expected *time.Time matches what the handler received.
func mustTime(t *testing.T, s string) *time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return &v
}

// registerAuth wires the auth boundary: known tokens resolve to a caller, an
// unknown token errors. Marked Maybe() so cases that don't authenticate (no/bad
// token) don't fail the mock.
func registerAuth(svc *service_mock.ServerService) {
	adminCtx := service.WithCaller(context.Background(), service.Caller{
		UserID: 1, UserName: "admin", IsAdmin: true, Source: service.AuditSourceApi,
	})
	userCtx := service.WithCaller(context.Background(), service.Caller{
		UserID: 2, UserName: "user1", IsAdmin: false, Source: service.AuditSourceApi,
	})
	svc.On("AuthContextByAPIToken", mock.Anything, AdminToken).Return(adminCtx, nil).Maybe()
	svc.On("AuthContextByAPIToken", mock.Anything, User1Token).Return(userCtx, nil).Maybe()
	svc.On("AuthContextByAPIToken", mock.Anything, "invalid").Return(nil, service.Unauthorized()).Maybe()
}

// requireJSONEqual compares two values at the JSON level, ignoring field
// ordering and Go-level type differences (e.g. *T vs T).
func requireJSONEqual(t *testing.T, want, got any) {
	t.Helper()
	wantJSON, err := json.Marshal(want)
	require.NoError(t, err)
	gotJSON, err := json.Marshal(got)
	require.NoError(t, err)

	var wantAny, gotAny any
	require.NoError(t, json.Unmarshal(wantJSON, &wantAny))
	require.NoError(t, json.Unmarshal(gotJSON, &gotAny))
	require.Equal(t, wantAny, gotAny)
}

// dispatch calls the client method matching the input type. A nil input means
// ProfileGet (its only method without an input).
func dispatch(ctx context.Context, c *remotesvc.Service, in any) (any, error) {
	switch p := in.(type) {
	case nil:
		return c.ProfileGet(ctx)

	// Payloads
	case service.PayloadsCreateInput:
		return c.PayloadsCreate(ctx, p)
	case service.PayloadsListInput:
		return c.PayloadsList(ctx, p)
	case service.PayloadsUpdateInput:
		return c.PayloadsUpdate(ctx, p)
	case service.PayloadsDeleteInput:
		return c.PayloadsDelete(ctx, p)
	case service.PayloadsClearInput:
		return c.PayloadsClear(ctx, p)

	// DNS records
	case service.DNSRecordsCreateInput:
		return c.DNSRecordsCreate(ctx, p)
	case service.DNSRecordsListInput:
		return c.DNSRecordsList(ctx, p)
	case service.DNSRecordsDeleteInput:
		return c.DNSRecordsDelete(ctx, p)
	case service.DNSRecordsClearInput:
		return c.DNSRecordsClear(ctx, p)

	// HTTP routes
	case service.HTTPRoutesCreateInput:
		return c.HTTPRoutesCreate(ctx, p)
	case service.HTTPRoutesListInput:
		return c.HTTPRoutesList(ctx, p)
	case service.HTTPRoutesUpdateInput:
		return c.HTTPRoutesUpdate(ctx, p)
	case service.HTTPRoutesDeleteInput:
		return c.HTTPRoutesDelete(ctx, p)
	case service.HTTPRoutesClearInput:
		return c.HTTPRoutesClear(ctx, p)

	// Events
	case service.EventsListInput:
		return c.EventsList(ctx, p)
	case service.EventsGetInput:
		return c.EventsGet(ctx, p)

	// Users
	case service.UsersCreateInput:
		return c.UsersCreate(ctx, p)
	case service.UsersDeleteInput:
		return c.UsersDelete(ctx, p)

	// Audit records
	case service.AuditRecordsListInput:
		return c.AuditRecordsList(ctx, p)
	case service.AuditRecordsGetInput:
		return c.AuditRecordsGet(ctx, p)

	default:
		panic(fmt.Sprintf("dispatch: unhandled input type %T", in))
	}
}

type rtCase struct {
	name  string
	token string

	// Mock setup. Empty svcMethod means no service call is expected (the handler
	// short-circuits, e.g. on a non-admin caller or a bad token).
	svcMethod  string
	svcNoInput bool // method takes only context (ProfileGet)
	svcInput   any  // also selects the client method via dispatch (nil => ProfileGet)
	svcResult  any
	svcErr     error

	// Expectation. wantKind nil => expect success, compare client output to
	// svcResult; otherwise expect a service.Error of that kind.
	wantKind     *service.ErrorKind
	wantMsg      string
	wantProblems map[string]string
}

func TestClientRoundTrip(t *testing.T) {
	tests := []rtCase{
		//
		// Payloads
		//
		{
			name:      "payloads create",
			token:     User1Token,
			svcMethod: "PayloadsCreate",
			svcInput: service.PayloadsCreateInput{
				Name:            "test",
				NotifyProtocols: []service.ProtoCategory{service.ProtoCategoryDns, service.ProtoCategorySmtp},
				StoreEvents:     true,
			},
			svcResult: &service.PayloadsCreateOutput{
				Name:            "test",
				Subdomain:       "abcd1234",
				NotifyProtocols: []service.ProtoCategory{service.ProtoCategoryDns, service.ProtoCategorySmtp},
				StoreEvents:     true,
			},
		},
		{
			name:      "payloads list",
			token:     User1Token,
			svcMethod: "PayloadsList",
			svcInput:  service.PayloadsListInput{Name: "foo", Page: 2, PerPage: 20},
			svcResult: service.PayloadsListOutput{{Name: "foo"}},
		},
		{
			name:      "payloads update",
			token:     User1Token,
			svcMethod: "PayloadsUpdate",
			svcInput: service.PayloadsUpdateInput{
				Name:            "payload1",
				NewName:         "new",
				NotifyProtocols: []service.ProtoCategory{service.ProtoCategorySmtp},
				StoreEvents:     new(false),
			},
			svcResult: &service.PayloadsUpdateOutput{Name: "new"},
		},
		{
			name:      "payloads delete",
			token:     User1Token,
			svcMethod: "PayloadsDelete",
			svcInput:  service.PayloadsDeleteInput{Name: "payload1"},
			svcResult: &service.PayloadsDeleteOutput{Name: "payload1"},
		},
		{
			name:      "payloads clear",
			token:     User1Token,
			svcMethod: "PayloadsClear",
			svcInput:  service.PayloadsClearInput{Name: "foo"},
			svcResult: service.PayloadsClearOutput{{Name: "foo"}},
		},

		//
		// DNS records
		//
		{
			name:      "dns records create",
			token:     User1Token,
			svcMethod: "DNSRecordsCreate",
			svcInput: service.DNSRecordsCreateInput{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         100,
				Type:        service.DNSRecordTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    service.DNSRecordStrategyAll,
			},
			svcResult: &service.DNSRecordsCreateOutput{Index: 1, Name: "test", Type: service.DNSRecordTypeA},
		},
		{
			name:      "dns records list",
			token:     User1Token,
			svcMethod: "DNSRecordsList",
			svcInput:  service.DNSRecordsListInput{PayloadName: "payload1"},
			svcResult: service.DNSRecordsListOutput{{Index: 1, Name: "test-a"}},
		},
		{
			name:      "dns records delete",
			token:     User1Token,
			svcMethod: "DNSRecordsDelete",
			svcInput:  service.DNSRecordsDeleteInput{PayloadName: "payload1", Index: 1},
			svcResult: &service.DNSRecordsDeleteOutput{Index: 1, Name: "test-a"},
		},
		{
			name:      "dns records clear",
			token:     User1Token,
			svcMethod: "DNSRecordsClear",
			svcInput:  service.DNSRecordsClearInput{PayloadName: "payload1", Name: "www"},
			svcResult: service.DNSRecordsClearOutput{{Index: 1}},
		},

		//
		// HTTP routes
		//
		{
			name:      "http routes create",
			token:     User1Token,
			svcMethod: "HTTPRoutesCreate",
			svcInput: service.HTTPRoutesCreateInput{
				PayloadName: "payload1",
				Method:      service.HTTPMethodGET,
				Path:        "/test",
				Code:        200,
				Headers:     map[string][]string{"Test": {"test"}},
				Body:        "dGVzdA==",
				IsDynamic:   true,
			},
			svcResult: &service.HTTPRoutesCreateOutput{Index: 1, Method: service.HTTPMethodGET, Path: "/test", Code: 200},
		},
		{
			name:      "http routes update",
			token:     User1Token,
			svcMethod: "HTTPRoutesUpdate",
			svcInput: service.HTTPRoutesUpdateInput{
				Payload:   "payload1",
				Index:     1,
				Method:    new(service.HTTPMethodPOST),
				Path:      new("/test2"),
				Code:      new(301),
				Headers:   map[string][]string{"X": {"x"}},
				Body:      new("MTIzNAo="),
				IsDynamic: new(false),
			},
			svcResult: &service.HTTPRoutesUpdateOutput{Index: 1, Method: service.HTTPMethodPOST, Path: "/test2", Code: 301},
		},
		{
			name:      "http routes list",
			token:     User1Token,
			svcMethod: "HTTPRoutesList",
			svcInput:  service.HTTPRoutesListInput{PayloadName: "payload1"},
			svcResult: service.HTTPRoutesListOutput{{Index: 1, Path: "/get"}},
		},
		{
			name:      "http routes delete",
			token:     User1Token,
			svcMethod: "HTTPRoutesDelete",
			svcInput:  service.HTTPRoutesDeleteInput{PayloadName: "payload1", Index: 1},
			svcResult: &service.HTTPRoutesDeleteOutput{Index: 1, Path: "/get"},
		},
		{
			name:      "http routes clear",
			token:     User1Token,
			svcMethod: "HTTPRoutesClear",
			svcInput:  service.HTTPRoutesClearInput{PayloadName: "payload1", Path: "/x"},
			svcResult: service.HTTPRoutesClearOutput{{Index: 1}},
		},

		//
		// Events
		//
		{
			name:      "events list",
			token:     User1Token,
			svcMethod: "EventsList",
			svcInput:  service.EventsListInput{PayloadName: "payload1", Limit: 5, Offset: 2},
			svcResult: service.EventsListOutput{{Index: 1, Protocol: service.EventProtocolHttp}},
		},
		{
			name:      "events get",
			token:     User1Token,
			svcMethod: "EventsGet",
			svcInput:  service.EventsGetInput{PayloadName: "payload1", Index: 2},
			svcResult: &service.EventsGetOutput{Index: 2, Protocol: service.EventProtocolHttp},
		},

		//
		// Profile
		//
		{
			name:       "profile get",
			token:      User1Token,
			svcMethod:  "ProfileGet",
			svcNoInput: true,
			svcInput:   nil,
			svcResult:  &service.ProfileGetOutput{Name: "user1"},
		},

		//
		// Users (admin only)
		//
		{
			name:      "users create",
			token:     AdminToken,
			svcMethod: "UsersCreate",
			svcInput: service.UsersCreateInput{
				Name:       "test",
				APIToken:   new("token"),
				TelegramID: new(int64(1234)),
			},
			svcResult: &service.UsersCreateOutput{Name: "test"},
		},
		{
			name:      "users delete",
			token:     AdminToken,
			svcMethod: "UsersDelete",
			svcInput:  service.UsersDeleteInput{Name: "user1"},
			svcResult: &service.UsersDeleteOutput{Name: "user1"},
		},

		//
		// Audit records (admin only)
		//
		{
			name:      "audit records list",
			token:     AdminToken,
			svcMethod: "AuditRecordsList",
			svcInput: service.AuditRecordsListInput{
				ActorID:      new(int64(1)),
				ActorName:    "user1",
				ResourceType: service.AuditResourceTypePayload,
				Action:       service.AuditActionCreate,
				From:         mustTime(t, "2026-01-01T09:00:00Z"),
				To:           mustTime(t, "2026-01-01T10:30:00Z"),
				Page:         1,
				PerPage:      1,
			},
			svcResult: service.AuditRecordsListOutput{{ID: 1, Action: service.AuditActionCreate}},
		},
		{
			name:      "audit records get",
			token:     AdminToken,
			svcMethod: "AuditRecordsGet",
			svcInput:  service.AuditRecordsGetInput{ID: 2},
			svcResult: &service.AuditRecordsGetOutput{ID: 2, Action: service.AuditActionUpdate},
		},

		//
		// Error round-trips (server maps service.Error/middleware -> status,
		// client maps status -> service.Error)
		//
		{
			name:      "not found",
			token:     User1Token,
			svcMethod: "PayloadsDelete",
			svcInput:  service.PayloadsDeleteInput{Name: "nope"},
			svcErr:    service.NotFoundf("payload not found"),
			wantKind:  new(service.ErrorKindNotFound),
			wantMsg:   "payload not found",
		},
		{
			name:      "conflict",
			token:     User1Token,
			svcMethod: "PayloadsCreate",
			svcInput:  service.PayloadsCreateInput{Name: "payload1"},
			svcErr:    service.Conflictf("payload already exists"),
			wantKind:  new(service.ErrorKindConflict),
			wantMsg:   "payload already exists",
		},
		{
			name:         "validation",
			token:        User1Token,
			svcMethod:    "PayloadsCreate",
			svcInput:     service.PayloadsCreateInput{Name: ""},
			svcErr:       service.Validation(map[string]string{"name": "cannot be blank"}),
			wantKind:     new(service.ErrorKindValidation),
			wantMsg:      "validation failed",
			wantProblems: map[string]string{"name": "cannot be blank"},
		},
		{
			name:      "internal",
			token:     User1Token,
			svcMethod: "PayloadsList",
			svcInput:  service.PayloadsListInput{},
			svcErr:    fmt.Errorf("boom"),
			wantKind:  new(service.ErrorKindInternal),
			wantMsg:   "internal error",
		},
		{
			// Real checkIsAdmin rejects a non-admin caller before the service.
			name:     "forbidden non-admin",
			token:    User1Token,
			svcInput: service.UsersCreateInput{Name: "x"},
			wantKind: new(service.ErrorKindForbidden),
			wantMsg:  "Admin only",
		},
		{
			// Real checkAuth rejects an unknown token before the service.
			name:     "unauthorized",
			token:    "invalid",
			svcInput: nil, // ProfileGet
			wantKind: new(service.ErrorKindUnauthorized),
			wantMsg:  "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service_mock.ServerService{}
			registerAuth(svc)

			if tt.svcMethod != "" {
				args := []any{mock.Anything}
				if !tt.svcNoInput {
					args = append(args, tt.svcInput)
				}
				svc.On(tt.svcMethod, args...).Return(tt.svcResult, tt.svcErr)
			}

			api, err := api.New(&api.Config{}, slog.New(slog.DiscardHandler), nil, svc)
			require.NoError(t, err)

			srv := httptest.NewServer(api.Handler())
			defer srv.Close()

			client := remotesvc.New(srv.URL, tt.token, srv.Client())

			got, err := dispatch(context.Background(), client, tt.svcInput)

			if tt.wantKind != nil {
				require.Error(t, err)
				se, ok := err.(service.Error)
				require.Truef(t, ok, "want service.Error, got %T: %v", err, err)
				require.Equal(t, *tt.wantKind, se.Kind)
				if tt.wantMsg != "" {
					require.Contains(t, se.Message, tt.wantMsg)
				}
				if tt.wantProblems != nil {
					require.Equal(t, tt.wantProblems, se.Problems)
				}
				return
			}

			require.NoError(t, err)
			requireJSONEqual(t, tt.svcResult, got)

			if tt.svcMethod != "" {
				args := []any{mock.Anything}
				if !tt.svcNoInput {
					args = append(args, tt.svcInput)
				}
				svc.AssertCalled(t, tt.svcMethod, args...)
			}
		})
	}
}
