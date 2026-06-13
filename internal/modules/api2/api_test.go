package api2_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/modules/api2"
	"github.com/nt0xa/sonar/internal/service"
	service_mock "github.com/nt0xa/sonar/internal/service/mock"
)

const (
	AdminToken = "admin-token"
	User1Token = "user1-token"
)

// mustTime parses an RFC3339 timestamp the same way the handlers do, so the
// *time.Time in an expected input matches what AuditRecordsList received.
func mustTime(t *testing.T, s string) *time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return &v
}

// registerAuth wires the auth boundary: each known token resolves to a caller
// (admin or regular), an unknown token errors. Marked Maybe() so cases that
// don't authenticate (missing token) don't fail.
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

type testCase struct {
	name   string
	method string
	path   string
	query  string
	token  string
	body   string

	// Expected service call. Empty svcMethod means no service call is expected
	// (auth/admin/decode/param failures short-circuit before the service).
	svcMethod  string
	svcNoInput bool // method takes only context (ProfileGet)
	svcInput   any
	svcResult  any
	svcErr     error

	status       int
	wantBody     any               // expected result body, compared via JSON round-trip
	wantMsg      string            // substring expected in the error "message"
	wantProblems map[string]string // expected "problems" for validation errors
}

func TestAPI(t *testing.T) {
	tests := []testCase{
		//
		// Payloads
		//
		{
			name:   "payloads create",
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			body:   `{"name":"test","notifyProtocols":["dns","smtp"],"storeEvents":true}`,
			svcMethod: "PayloadsCreate",
			svcInput: service.PayloadsCreateInput{
				Name:            "test",
				NotifyProtocols: []service.ProtoCategory{service.ProtoCategoryDns, service.ProtoCategorySmtp},
				StoreEvents:     true,
			},
			svcResult: &service.Payload{
				Name:            "test",
				Subdomain:       "abcd1234",
				NotifyProtocols: []service.ProtoCategory{service.ProtoCategoryDns, service.ProtoCategorySmtp},
				StoreEvents:     true,
			},
			status:   http.StatusCreated,
			wantBody: &service.Payload{Name: "test", Subdomain: "abcd1234", NotifyProtocols: []service.ProtoCategory{service.ProtoCategoryDns, service.ProtoCategorySmtp}, StoreEvents: true},
		},
		{
			name:   "payloads list",
			method: "GET",
			path:   "/payloads",
			query:  "name=foo&page=2&perPage=20",
			token:  User1Token,
			svcMethod: "PayloadsList",
			svcInput:  service.PayloadsListInput{Name: "foo", Page: 2, PerPage: 20},
			svcResult: []service.Payload{{Name: "foo"}},
			status:    http.StatusOK,
			wantBody:  []service.Payload{{Name: "foo"}},
		},
		{
			name:   "payloads update",
			method: "PATCH",
			path:   "/payloads/payload1",
			token:  User1Token,
			body:   `{"name":"new","notifyProtocols":["smtp"],"storeEvents":false}`,
			svcMethod: "PayloadsUpdate",
			svcInput: service.PayloadsUpdateInput{
				Name:            "payload1",
				NewName:         "new",
				NotifyProtocols: []service.ProtoCategory{service.ProtoCategorySmtp},
				StoreEvents:     new(false),
			},
			svcResult: &service.Payload{Name: "new"},
			status:    http.StatusOK,
			wantBody:  &service.Payload{Name: "new"},
		},
		{
			name:   "payloads delete",
			method: "DELETE",
			path:   "/payloads/payload1",
			token:  User1Token,
			svcMethod: "PayloadsDelete",
			svcInput:  service.PayloadsDeleteInput{Name: "payload1"},
			svcResult: &service.Payload{Name: "payload1"},
			status:    http.StatusOK,
			wantBody:  &service.Payload{Name: "payload1"},
		},
		{
			name:   "payloads clear",
			method: "DELETE",
			path:   "/payloads",
			query:  "name=foo",
			token:  User1Token,
			svcMethod: "PayloadsClear",
			svcInput:  service.PayloadsClearInput{Name: "foo"},
			svcResult: []service.Payload{{Name: "foo"}},
			status:    http.StatusOK,
			wantBody:  []service.Payload{{Name: "foo"}},
		},

		//
		// DNS records
		//
		{
			name:   "dns records create",
			method: "POST",
			path:   "/dns-records",
			token:  User1Token,
			body:   `{"payloadName":"payload1","name":"test","type":"A","ttl":100,"values":["127.0.0.1"],"strategy":"all"}`,
			svcMethod: "DNSRecordsCreate",
			svcInput: service.DNSRecordsCreateInput{
				PayloadName: "payload1",
				Name:        "test",
				TTL:         100,
				Type:        service.DNSRecordTypeA,
				Values:      []string{"127.0.0.1"},
				Strategy:    service.DNSRecordStrategyAll,
			},
			svcResult: &service.DNSRecord{Index: 1, Name: "test", Type: service.DNSRecordTypeA},
			status:    http.StatusCreated,
			wantBody:  &service.DNSRecord{Index: 1, Name: "test", Type: service.DNSRecordTypeA},
		},
		{
			name:   "dns records list",
			method: "GET",
			path:   "/dns-records/payload1",
			token:  User1Token,
			svcMethod: "DNSRecordsList",
			svcInput:  service.DNSRecordsListInput{PayloadName: "payload1"},
			svcResult: []service.DNSRecord{{Index: 1, Name: "test-a"}},
			status:    http.StatusOK,
			wantBody:  []service.DNSRecord{{Index: 1, Name: "test-a"}},
		},
		{
			name:   "dns records delete",
			method: "DELETE",
			path:   "/dns-records/payload1/1",
			token:  User1Token,
			svcMethod: "DNSRecordsDelete",
			svcInput:  service.DNSRecordsDeleteInput{PayloadName: "payload1", Index: 1},
			svcResult: &service.DNSRecord{Index: 1, Name: "test-a"},
			status:    http.StatusOK,
			wantBody:  &service.DNSRecord{Index: 1, Name: "test-a"},
		},
		{
			name:   "dns records clear",
			method: "DELETE",
			path:   "/dns-records/payload1",
			query:  "name=www",
			token:  User1Token,
			svcMethod: "DNSRecordsClear",
			svcInput:  service.DNSRecordsClearInput{PayloadName: "payload1", Name: "www"},
			svcResult: []service.DNSRecord{{Index: 1}},
			status:    http.StatusOK,
			wantBody:  []service.DNSRecord{{Index: 1}},
		},

		//
		// HTTP routes
		//
		{
			name:   "http routes create",
			method: "POST",
			path:   "/http-routes",
			token:  User1Token,
			body:   `{"payloadName":"payload1","method":"GET","path":"/test","code":200,"headers":{"Test":["test"]},"body":"dGVzdA==","isDynamic":true}`,
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
			svcResult: &service.HTTPRoute{Index: 1, Method: service.HTTPMethodGET, Path: "/test", Code: 200},
			status:    http.StatusCreated,
			wantBody:  &service.HTTPRoute{Index: 1, Method: service.HTTPMethodGET, Path: "/test", Code: 200},
		},
		{
			name:   "http routes update",
			method: "PATCH",
			path:   "/http-routes/payload1/1",
			token:  User1Token,
			body:   `{"method":"POST","path":"/test2","code":301,"headers":{"X":["x"]},"body":"MTIzNAo=","isDynamic":false}`,
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
			svcResult: &service.HTTPRoute{Index: 1, Method: service.HTTPMethodPOST, Path: "/test2", Code: 301},
			status:    http.StatusOK,
			wantBody:  &service.HTTPRoute{Index: 1, Method: service.HTTPMethodPOST, Path: "/test2", Code: 301},
		},
		{
			name:   "http routes list",
			method: "GET",
			path:   "/http-routes/payload1",
			token:  User1Token,
			svcMethod: "HTTPRoutesList",
			svcInput:  service.HTTPRoutesListInput{PayloadName: "payload1"},
			svcResult: []service.HTTPRoute{{Index: 1, Path: "/get"}},
			status:    http.StatusOK,
			wantBody:  []service.HTTPRoute{{Index: 1, Path: "/get"}},
		},
		{
			name:   "http routes delete",
			method: "DELETE",
			path:   "/http-routes/payload1/1",
			token:  User1Token,
			svcMethod: "HTTPRoutesDelete",
			svcInput:  service.HTTPRoutesDeleteInput{PayloadName: "payload1", Index: 1},
			svcResult: &service.HTTPRoute{Index: 1, Path: "/get"},
			status:    http.StatusOK,
			wantBody:  &service.HTTPRoute{Index: 1, Path: "/get"},
		},
		{
			name:   "http routes clear",
			method: "DELETE",
			path:   "/http-routes/payload1",
			query:  "path=/x",
			token:  User1Token,
			svcMethod: "HTTPRoutesClear",
			svcInput:  service.HTTPRoutesClearInput{PayloadName: "payload1", Path: "/x"},
			svcResult: []service.HTTPRoute{{Index: 1}},
			status:    http.StatusOK,
			wantBody:  []service.HTTPRoute{{Index: 1}},
		},

		//
		// Events
		//
		{
			name:   "events list",
			method: "GET",
			path:   "/events/payload1",
			query:  "limit=5&offset=2",
			token:  User1Token,
			svcMethod: "EventsList",
			svcInput:  service.EventsListInput{PayloadName: "payload1", Limit: 5, Offset: 2},
			svcResult: []service.Event{{Index: 1, Protocol: service.EventProtocolHttp}},
			status:    http.StatusOK,
			wantBody:  []service.Event{{Index: 1, Protocol: service.EventProtocolHttp}},
		},
		{
			name:   "events get",
			method: "GET",
			path:   "/events/payload1/2",
			token:  User1Token,
			svcMethod: "EventsGet",
			svcInput:  service.EventsGetInput{PayloadName: "payload1", Index: 2},
			svcResult: &service.Event{Index: 2, Protocol: service.EventProtocolHttp},
			status:    http.StatusOK,
			wantBody:  &service.Event{Index: 2, Protocol: service.EventProtocolHttp},
		},

		//
		// Profile
		//
		{
			name:   "profile get",
			method: "GET",
			path:   "/profile",
			token:  User1Token,
			svcMethod:  "ProfileGet",
			svcNoInput: true,
			svcResult:  &service.User{Name: "user1"},
			status:     http.StatusOK,
			wantBody:   &service.User{Name: "user1"},
		},

		//
		// Users (admin only)
		//
		{
			name:   "users create",
			method: "POST",
			path:   "/users",
			token:  AdminToken,
			body:   `{"name":"test","apiToken":"token","telegramId":1234}`,
			svcMethod: "UsersCreate",
			svcInput: service.UsersCreateInput{
				Name:       "test",
				APIToken:   new("token"),
				TelegramID: new(int64(1234)),
			},
			svcResult: &service.User{Name: "test"},
			status:    http.StatusCreated,
			wantBody:  &service.User{Name: "test"},
		},
		{
			name:   "users delete",
			method: "DELETE",
			path:   "/users/user1",
			token:  AdminToken,
			svcMethod: "UsersDelete",
			svcInput:  service.UsersDeleteInput{Name: "user1"},
			svcResult: &service.User{Name: "user1"},
			status:    http.StatusOK,
			wantBody:  &service.User{Name: "user1"},
		},

		//
		// Audit records (admin only)
		//
		{
			name:   "audit records list",
			method: "GET",
			path:   "/audit-records",
			query:  "actorId=1&actorName=user1&resourceType=payload&action=create&page=1&perPage=1&from=2026-01-01T09:00:00Z&to=2026-01-01T10:30:00Z",
			token:  AdminToken,
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
			svcResult: []service.AuditRecord{{ID: 1, Action: service.AuditActionCreate}},
			status:    http.StatusOK,
			wantBody:  []service.AuditRecord{{ID: 1, Action: service.AuditActionCreate}},
		},
		{
			name:   "audit records get",
			method: "GET",
			path:   "/audit-records/2",
			token:  AdminToken,
			svcMethod: "AuditRecordsGet",
			svcInput:  service.AuditRecordsGetInput{ID: 2},
			svcResult: &service.AuditRecord{ID: 2, Action: service.AuditActionUpdate},
			status:    http.StatusOK,
			wantBody:  &service.AuditRecord{ID: 2, Action: service.AuditActionUpdate},
		},

		//
		// Error matrix (each mapped kind once)
		//
		{
			name:   "bad json body",
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			body:   `{`,
			status: http.StatusBadRequest,
		},
		{
			name:   "bad path param",
			method: "DELETE",
			path:   "/dns-records/payload1/abc",
			token:  User1Token,
			status:  http.StatusBadRequest,
			wantMsg: "integer",
		},
		{
			name:    "missing token",
			method:  "GET",
			path:    "/profile",
			status:  http.StatusUnauthorized,
			wantMsg: "Missing token",
		},
		{
			name:    "invalid token",
			method:  "GET",
			path:    "/profile",
			token:   "invalid",
			status:  http.StatusUnauthorized,
			wantMsg: "Invalid token",
		},
		{
			name:    "forbidden non-admin",
			method:  "POST",
			path:    "/users",
			token:   User1Token,
			body:    `{"name":"x"}`,
			status:  http.StatusForbidden,
			wantMsg: "Admin only",
		},
		{
			name:   "not found",
			method: "DELETE",
			path:   "/payloads/nope",
			token:  User1Token,
			svcMethod: "PayloadsDelete",
			svcInput:  service.PayloadsDeleteInput{Name: "nope"},
			svcErr:    service.NotFoundf("payload not found"),
			status:    http.StatusNotFound,
			wantMsg:   "not found",
		},
		{
			name:   "conflict",
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			body:   `{"name":"payload1"}`,
			svcMethod: "PayloadsCreate",
			svcInput:  service.PayloadsCreateInput{Name: "payload1"},
			svcErr:    service.Conflictf("payload already exists"),
			status:    http.StatusConflict,
			wantMsg:   "already exists",
		},
		{
			name:   "validation",
			method: "POST",
			path:   "/payloads",
			token:  User1Token,
			body:   `{"name":""}`,
			svcMethod:    "PayloadsCreate",
			svcInput:     service.PayloadsCreateInput{Name: ""},
			svcErr:       service.Validation(map[string]string{"name": "cannot be blank"}),
			status:       http.StatusUnprocessableEntity,
			wantMsg:      "validation failed",
			wantProblems: map[string]string{"name": "cannot be blank"},
		},
		{
			name:   "internal error",
			method: "GET",
			path:   "/payloads",
			token:  User1Token,
			svcMethod: "PayloadsList",
			svcInput:  service.PayloadsListInput{},
			svcErr:    fmt.Errorf("boom"),
			status:    http.StatusInternalServerError,
			wantMsg:   "internal error",
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

			api, err := api2.New(&api2.Config{Admin: AdminToken}, slog.New(slog.DiscardHandler), svc)
			require.NoError(t, err)

			srv := httptest.NewServer(api.Handler())
			defer srv.Close()

			resp := doRequest(t, srv, tt)
			defer resp.Body.Close()

			require.Equal(t, tt.status, resp.StatusCode)

			raw, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.wantBody != nil {
				assertJSONEqual(t, tt.wantBody, raw)
			}

			if tt.wantMsg != "" || tt.wantProblems != nil {
				var e struct {
					Message  string            `json:"message"`
					Problems map[string]string `json:"problems"`
				}
				require.NoError(t, json.Unmarshal(raw, &e))
				if tt.wantMsg != "" {
					require.Contains(t, e.Message, tt.wantMsg)
				}
				if tt.wantProblems != nil {
					require.Equal(t, tt.wantProblems, e.Problems)
				}
			}

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

func doRequest(t *testing.T, srv *httptest.Server, tt testCase) *http.Response {
	t.Helper()

	var body io.Reader
	if tt.body != "" {
		body = strings.NewReader(tt.body)
	}

	url := srv.URL + tt.path
	if tt.query != "" {
		url += "?" + tt.query
	}

	req, err := http.NewRequest(tt.method, url, body)
	require.NoError(t, err)

	if tt.token != "" {
		req.Header.Set("Authorization", "Bearer "+tt.token)
	}
	if tt.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)

	return resp
}

// assertJSONEqual compares want (re-encoded to JSON) with the raw response body
// at the structural level, ignoring field ordering and formatting.
func assertJSONEqual(t *testing.T, want any, raw []byte) {
	t.Helper()

	wantJSON, err := json.Marshal(want)
	require.NoError(t, err)

	var wantAny, gotAny any
	require.NoError(t, json.Unmarshal(wantJSON, &wantAny))
	require.NoError(t, json.Unmarshal(raw, &gotAny))

	require.Equal(t, wantAny, gotAny)
}
