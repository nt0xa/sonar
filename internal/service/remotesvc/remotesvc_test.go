package remotesvc_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/internal/service/remotesvc"
)

// recordingServer captures the last request it received and replies with a
// fixed status and raw body.
type recordingServer struct {
	srv *httptest.Server

	method string
	path   string
	query  string
	auth   string
	body   []byte

	status int
	reply  string
}

func newServer(t *testing.T) *recordingServer {
	t.Helper()
	rs := &recordingServer{status: http.StatusOK, reply: "null"}
	rs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.method = r.Method
		rs.path = r.URL.Path
		rs.query = r.URL.RawQuery
		rs.auth = r.Header.Get("Authorization")
		rs.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rs.status)
		_, _ = io.WriteString(w, rs.reply)
	}))
	t.Cleanup(rs.srv.Close)
	return rs
}

func (rs *recordingServer) client() *remotesvc.Service {
	return remotesvc.New(rs.srv.URL, "secret-token", rs.srv.Client())
}

func TestPayloadsCreate(t *testing.T) {
	rs := newServer(t)
	rs.reply = `{"name":"foo","subdomain":"abc","storeEvents":true}`

	out, err := rs.client().PayloadsCreate(context.Background(), service.PayloadsCreateInput{
		Name:            "foo",
		NotifyProtocols: []service.ProtoCategory{service.ProtoCategoryDns},
		StoreEvents:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rs.method != http.MethodPost || rs.path != "/payloads" {
		t.Errorf("got %s %s, want POST /payloads", rs.method, rs.path)
	}
	if rs.auth != "Bearer secret-token" {
		t.Errorf("auth header = %q", rs.auth)
	}

	var sent map[string]any
	if err := json.Unmarshal(rs.body, &sent); err != nil {
		t.Fatalf("request body not JSON: %v", err)
	}
	if sent["name"] != "foo" || sent["storeEvents"] != true {
		t.Errorf("unexpected request body: %s", rs.body)
	}

	if out.Name != "foo" || out.Subdomain != "abc" || !out.StoreEvents {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestPayloadsList_Query(t *testing.T) {
	rs := newServer(t)
	rs.reply = `[{"name":"foo"}]`

	out, err := rs.client().PayloadsList(context.Background(), service.PayloadsListInput{
		Name:    "foo",
		Page:    2,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rs.method != http.MethodGet || rs.path != "/payloads" {
		t.Errorf("got %s %s, want GET /payloads", rs.method, rs.path)
	}
	if rs.query != "name=foo&page=2&perPage=10" {
		t.Errorf("query = %q", rs.query)
	}
	if len(out) != 1 || out[0].Name != "foo" {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestEventsGet_PathParams(t *testing.T) {
	rs := newServer(t)
	rs.reply = `{"index":7,"uuid":"u"}`

	out, err := rs.client().EventsGet(context.Background(), service.EventsGetInput{
		PayloadName: "foo",
		Index:       7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rs.path != "/events/foo/7" {
		t.Errorf("path = %q, want /events/foo/7", rs.path)
	}
	if out.Index != 7 {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestDNSRecordsClear_Query(t *testing.T) {
	rs := newServer(t)
	rs.reply = `[]`

	_, err := rs.client().DNSRecordsClear(context.Background(), service.DNSRecordsClearInput{
		PayloadName: "foo",
		Name:        "bar",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rs.method != http.MethodDelete || rs.path != "/dns-records/foo" {
		t.Errorf("got %s %s, want DELETE /dns-records/foo", rs.method, rs.path)
	}
	if rs.query != "name=bar" {
		t.Errorf("query = %q", rs.query)
	}
}

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		reply    string
		wantKind service.ErrorKind
		wantMsg  string
		wantProb map[string]string
	}{
		{
			name:     "not found",
			status:   http.StatusNotFound,
			reply:    `{"message":"payload not found"}`,
			wantKind: service.ErrorKindNotFound,
			wantMsg:  "payload not found",
		},
		{
			name:     "conflict",
			status:   http.StatusConflict,
			reply:    `{"message":"already exists"}`,
			wantKind: service.ErrorKindConflict,
			wantMsg:  "already exists",
		},
		{
			name:     "validation",
			status:   http.StatusUnprocessableEntity,
			reply:    `{"message":"validation failed","problems":{"name":"is required"}}`,
			wantKind: service.ErrorKindValidation,
			wantMsg:  "validation failed",
			wantProb: map[string]string{"name": "is required"},
		},
		{
			name:     "unauthorized",
			status:   http.StatusUnauthorized,
			reply:    `{"message":"Invalid token"}`,
			wantKind: service.ErrorKindUnauthorized,
			wantMsg:  "Invalid token",
		},
		{
			name:     "forbidden",
			status:   http.StatusForbidden,
			reply:    `{"message":"Admin only"}`,
			wantKind: service.ErrorKindForbidden,
			wantMsg:  "Admin only",
		},
		{
			name:     "bad request",
			status:   http.StatusBadRequest,
			reply:    `{"message":"index must be an integer"}`,
			wantKind: service.ErrorKindBadRequest,
			wantMsg:  "index must be an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := newServer(t)
			rs.status = tt.status
			rs.reply = tt.reply

			_, err := rs.client().ProfileGet(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			se, ok := err.(service.Error)
			if !ok {
				t.Fatalf("error is %T, want service.Error", err)
			}
			if se.Kind != tt.wantKind {
				t.Errorf("kind = %d, want %d", se.Kind, tt.wantKind)
			}
			if se.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", se.Message, tt.wantMsg)
			}
			if tt.wantProb != nil {
				if len(se.Problems) != len(tt.wantProb) {
					t.Errorf("problems = %v, want %v", se.Problems, tt.wantProb)
				}
				for k, v := range tt.wantProb {
					if se.Problems[k] != v {
						t.Errorf("problems[%q] = %q, want %q", k, se.Problems[k], v)
					}
				}
			}
		})
	}
}
