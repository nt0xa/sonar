package templates

import (
	"testing"

	"github.com/nt0xa/sonar/internal/service"
)

// TestRenderResultAllOutputs renders every service output type to guard against
// template field/type-assertion regressions (e.g. sprig funcs like `upper`
// receiving a defined enum type instead of a string).
func TestRenderResultAllOutputs(t *testing.T) {
	tpl := New("example.com")

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
