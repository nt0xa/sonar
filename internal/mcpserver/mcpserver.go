// Package mcpserver exposes [service.Service] operations as MCP tools.
package mcpserver

import (
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nt0xa/sonar/internal/service"
)

// New returns an MCP server with a tool per service operation.
func New(svc service.Service, version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "sonar", Version: version}, nil)

	// Payloads

	addTool(s, "payloads_list", "List payloads", readOnly,
		func(ctx context.Context, in struct {
			Name    string `json:"name,omitempty" jsonschema:"Filter by payload name substring"`
			Page    uint   `json:"page,omitempty" jsonschema:"Page number, starting from 1"`
			PerPage uint   `json:"perPage,omitempty" jsonschema:"Payloads per page"`
		}) (any, error) {
			return svc.PayloadsList(ctx, service.PayloadsListInput(in))
		})

	addTool(s, "payloads_create", "Create a new payload (a unique subdomain that captures DNS/HTTP/SMTP/FTP interactions)", write,
		func(ctx context.Context, in struct {
			Name            string                  `json:"name" jsonschema:"Payload name"`
			NotifyProtocols []service.ProtoCategory `json:"notifyProtocols,omitempty" jsonschema:"Protocols to notify connected messengers (Telegram, Slack, Lark) about: dns, http, smtp, ftp (default none). Only affects messenger notifications; events are always stored"`
		}) (any, error) {
			if in.NotifyProtocols == nil {
				in.NotifyProtocols = []service.ProtoCategory{}
			}
			return svc.PayloadsCreate(ctx, service.PayloadsCreateInput(in))
		})

	addTool(s, "payloads_update", "Modify existing payload", write,
		func(ctx context.Context, in struct {
			Name            string                  `json:"name" jsonschema:"Payload name"`
			NewName         string                  `json:"newName,omitempty" jsonschema:"New payload name"`
			NotifyProtocols []service.ProtoCategory `json:"notifyProtocols,omitempty" jsonschema:"Protocols to notify connected messengers (Telegram, Slack, Lark) about: dns, http, smtp, ftp (omit to keep current, [] for none). Only affects messenger notifications; events are always stored"`
		}) (any, error) {
			return svc.PayloadsUpdate(ctx, service.PayloadsUpdateInput(in))
		})

	addTool(s, "payloads_delete", "Delete payload", destructive,
		func(ctx context.Context, in struct {
			Name string `json:"name" jsonschema:"Payload name"`
		}) (any, error) {
			return svc.PayloadsDelete(ctx, service.PayloadsDeleteInput(in))
		})

	addTool(s, "payloads_clear", "Delete multiple payloads", destructive,
		func(ctx context.Context, in struct {
			Name string `json:"name,omitempty" jsonschema:"Delete only payloads whose name contains this substring"`
		}) (any, error) {
			return svc.PayloadsClear(ctx, service.PayloadsClearInput(in))
		})

	// DNS records

	addTool(s, "dns_records_list", "List DNS records of a payload", readOnly,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
		}) (any, error) {
			return svc.DNSRecordsList(ctx, service.DNSRecordsListInput(in))
		})

	addTool(s, "dns_records_create", "Create a DNS record for a payload subdomain", write,
		func(ctx context.Context, in struct {
			PayloadName string                    `json:"payloadName" jsonschema:"Payload name"`
			Name        string                    `json:"name" jsonschema:"Subdomain of the payload subdomain, e.g. \"test\" or \"*\""`
			Values      []string                  `json:"values" jsonschema:"Record values"`
			Type        service.DNSRecordType     `json:"type,omitempty" jsonschema:"Record type: A, AAAA, MX, TXT, CNAME, NS, CAA (default A)"`
			TTL         int                       `json:"ttl,omitempty" jsonschema:"Record TTL in seconds (default 60)"`
			Strategy    service.DNSRecordStrategy `json:"strategy,omitempty" jsonschema:"Strategy for multiple values: all, round-robin, rebind (default all)"`
		}) (any, error) {
			return svc.DNSRecordsCreate(ctx, service.DNSRecordsCreateInput{
				PayloadName: in.PayloadName,
				Name:        in.Name,
				Values:      in.Values,
				Type:        cmp.Or(in.Type, service.DNSRecordTypeA),
				TTL:         cmp.Or(in.TTL, 60),
				Strategy:    cmp.Or(in.Strategy, service.DNSRecordStrategyAll),
			})
		})

	addTool(s, "dns_records_delete", "Delete DNS record by index", destructive,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
			Index       int64  `json:"index" jsonschema:"DNS record index"`
		}) (any, error) {
			return svc.DNSRecordsDelete(ctx, service.DNSRecordsDeleteInput(in))
		})

	addTool(s, "dns_records_clear", "Delete multiple DNS records of a payload", destructive,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
			Name        string `json:"name,omitempty" jsonschema:"Delete only records with this subdomain"`
		}) (any, error) {
			return svc.DNSRecordsClear(ctx, service.DNSRecordsClearInput(in))
		})

	// HTTP routes

	addTool(s, "http_routes_list", "List HTTP routes of a payload (bodies are base64-encoded)", readOnly,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
		}) (any, error) {
			return svc.HTTPRoutesList(ctx, service.HTTPRoutesListInput(in))
		})

	addTool(s, "http_routes_create", "Create an HTTP route that serves a custom response on the payload subdomain", write,
		func(ctx context.Context, in struct {
			PayloadName string              `json:"payloadName" jsonschema:"Payload name"`
			Method      service.HTTPMethod  `json:"method,omitempty" jsonschema:"Request method: GET, HEAD, POST, PUT, PATCH, DELETE, CONNECT, OPTIONS, TRACE, ANY (default GET)"`
			Path        string              `json:"path,omitempty" jsonschema:"Request path, must start with / (default /)"`
			Code        int                 `json:"code,omitempty" jsonschema:"Response status code (default 200)"`
			Headers     map[string][]string `json:"headers,omitempty" jsonschema:"Response headers"`
			Body        string              `json:"body,omitempty" jsonschema:"Response body (plain text)"`
			IsDynamic   bool                `json:"isDynamic,omitempty" jsonschema:"Interpret body and headers as Go templates"`
		}) (any, error) {
			return svc.HTTPRoutesCreate(ctx, service.HTTPRoutesCreateInput{
				PayloadName: in.PayloadName,
				Method:      cmp.Or(in.Method, service.HTTPMethodGET),
				Path:        cmp.Or(in.Path, "/"),
				Code:        cmp.Or(in.Code, 200),
				Headers:     in.Headers,
				Body:        base64.StdEncoding.EncodeToString([]byte(in.Body)),
				IsDynamic:   in.IsDynamic,
			})
		})

	addTool(s, "http_routes_update", "Update HTTP route by index; omitted fields are left unchanged", write,
		func(ctx context.Context, in struct {
			PayloadName string              `json:"payloadName" jsonschema:"Payload name"`
			Index       int64               `json:"index" jsonschema:"HTTP route index"`
			Method      *service.HTTPMethod `json:"method,omitempty" jsonschema:"Request method: GET, HEAD, POST, PUT, PATCH, DELETE, CONNECT, OPTIONS, TRACE, ANY"`
			Path        *string             `json:"path,omitempty" jsonschema:"Request path, must start with /"`
			Code        *int                `json:"code,omitempty" jsonschema:"Response status code"`
			Headers     map[string][]string `json:"headers,omitempty" jsonschema:"Response headers (replace existing)"`
			Body        *string             `json:"body,omitempty" jsonschema:"Response body (plain text)"`
			IsDynamic   *bool               `json:"isDynamic,omitempty" jsonschema:"Interpret body and headers as Go templates"`
		}) (any, error) {
			if in.Body != nil {
				b := base64.StdEncoding.EncodeToString([]byte(*in.Body))
				in.Body = &b
			}
			return svc.HTTPRoutesUpdate(ctx, service.HTTPRoutesUpdateInput{
				Payload:   in.PayloadName,
				Index:     in.Index,
				Method:    in.Method,
				Path:      in.Path,
				Code:      in.Code,
				Headers:   in.Headers,
				Body:      in.Body,
				IsDynamic: in.IsDynamic,
			})
		})

	addTool(s, "http_routes_delete", "Delete HTTP route by index", destructive,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
			Index       int64  `json:"index" jsonschema:"HTTP route index"`
		}) (any, error) {
			return svc.HTTPRoutesDelete(ctx, service.HTTPRoutesDeleteInput(in))
		})

	addTool(s, "http_routes_clear", "Delete multiple HTTP routes of a payload", destructive,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
			Path        string `json:"path,omitempty" jsonschema:"Delete only routes with this path"`
		}) (any, error) {
			return svc.HTTPRoutesClear(ctx, service.HTTPRoutesClearInput(in))
		})

	// Events

	addTool(s, "events_list", "List interactions (events) captured by a payload; newest first by default (raw data is base64-encoded)", readOnly,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
			Count       uint   `json:"count,omitempty" jsonschema:"Number of events, max 100 (default 10)"`
			After       *int64 `json:"after,omitempty" jsonschema:"Return events after this index, oldest first"`
			Before      *int64 `json:"before,omitempty" jsonschema:"Return events before this index, newest first"`
		}) (any, error) {
			return svc.EventsList(ctx, service.EventsListInput{
				PayloadName: in.PayloadName,
				Count:       in.Count,
				After:       in.After,
				Before:      in.Before,
			})
		})

	addTool(s, "events_get", "Get payload event by index (raw data is base64-encoded)", readOnly,
		func(ctx context.Context, in struct {
			PayloadName string `json:"payloadName" jsonschema:"Payload name"`
			Index       int64  `json:"index" jsonschema:"Event index"`
		}) (any, error) {
			return svc.EventsGet(ctx, service.EventsGetInput(in))
		})

	return s
}

var (
	readOnly    = &mcp.ToolAnnotations{ReadOnlyHint: true}
	write       = &mcp.ToolAnnotations{DestructiveHint: new(false)}
	destructive = &mcp.ToolAnnotations{DestructiveHint: new(true)}
)

// addTool registers a tool that returns the service output as JSON text.
func addTool[In any](s *mcp.Server, name, desc string, ann *mcp.ToolAnnotations, f func(context.Context, In) (any, error)) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc, Annotations: ann},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			out, err := f(ctx, in)
			if err != nil {
				return nil, nil, err
			}
			b, err := json.Marshal(out)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
		})
}
