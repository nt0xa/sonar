package templates_test

import (
	"strings"
	"testing"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/internal/templates"
)

// TestRenderResultAllOutputs renders every service output type to guard against
// template field/type-assertion regressions (e.g. sprig funcs like `upper`
// receiving a defined enum type instead of a string).
func TestRenderResultAllOutputs(t *testing.T) {
	tpl := templates.New("example.com")

	cases := []any{
		&service.ProfileGetOutput{Name: "user"},

		&service.PayloadsCreateOutput{Name: "p"},
		&service.PayloadsUpdateOutput{Name: "p"},
		service.PayloadsListOutput{{Name: "p"}},
		&service.PayloadsDeleteOutput{Name: "p"},
		service.PayloadsClearOutput{{Name: "p"}},

		&service.DNSRecordsCreateOutput{Index: 1, Type: service.DNSRecordTypeA, Values: []string{"1.2.3.4"}},
		service.DNSRecordsListOutput{{Index: 1, Type: service.DNSRecordTypeA, Values: []string{"1.2.3.4"}}},
		&service.DNSRecordsDeleteOutput{Index: 1},
		service.DNSRecordsClearOutput{{Index: 1}},

		&service.HTTPRoutesCreateOutput{Index: 1, Method: service.HTTPMethodGET},
		&service.HTTPRoutesUpdateOutput{Index: 1, Method: service.HTTPMethodGET},
		service.HTTPRoutesListOutput{{Index: 1, Method: service.HTTPMethodGET}},
		&service.HTTPRoutesDeleteOutput{Index: 1},
		service.HTTPRoutesClearOutput{{Index: 1}},

		&service.UsersCreateOutput{Name: "user"},
		&service.UsersDeleteOutput{Name: "user"},
		service.UsersListOutput{{Name: "user", IsAdmin: true, TelegramID: new(int64(1))}},

		&service.EventsGetOutput{Index: 1, Protocol: service.EventProtocolHttp},
		service.EventsListOutput{{Index: 1, Protocol: service.EventProtocolDns}},

		service.AuditRecordsListOutput{{ID: 1, Action: service.AuditActionCreate}},
		&service.AuditRecordsGetOutput{ID: 1, Action: service.AuditActionCreate},
	}

	for _, c := range cases {
		if _, err := tpl.RenderResult(c); err != nil {
			t.Errorf("RenderResult(%T): %v", c, err)
		}
	}
}

func TestRenderNotificationBatch(t *testing.T) {
	tpl := templates.New("example.com")

	payload := &database.Payload{Name: "p"}
	ns := []*modules.Notification{
		{Payload: payload, Event: &database.Event{Protocol: "dns"}},
		{Payload: payload, Event: &database.Event{Protocol: "http"}},
		{Payload: payload, Event: &database.Event{Protocol: "dns"}},
	}

	got, err := tpl.RenderNotificationBatch(ns)
	if err != nil {
		t.Fatal(err)
	}

	want := "DNS 2 · HTTP 1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderAuditRecordActor(t *testing.T) {
	tpl := templates.New("example.com")

	cases := []struct {
		rec  service.AuditRecord
		want string
	}{
		{service.AuditRecord{ID: 1, Action: service.AuditActionCreate, ResourceType: service.AuditResourceTypePayload, Source: service.AuditSourceApi, ActorName: "user1"},
			"CREATE payload via api by user1 01 Jan 0001"},
		{service.AuditRecord{ID: 1, Action: service.AuditActionCreate, ResourceType: service.AuditResourceTypePayload, Source: service.AuditSourceApi},
			"CREATE payload via api by unknown 01 Jan 0001"},
		{service.AuditRecord{ID: 1, Action: service.AuditActionCreate, ResourceType: service.AuditResourceTypeEvent, Source: service.AuditSourceSystem},
			"CREATE event via system 01 Jan 0001"},
	}

	for _, c := range cases {
		got, err := tpl.RenderResult(service.AuditRecordsListOutput{c.rec})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, c.want) {
			t.Errorf("got %q, want it to contain %q", got, c.want)
		}
	}
}
