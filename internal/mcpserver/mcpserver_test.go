package mcpserver_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/mcpserver"
	"github.com/nt0xa/sonar/internal/service"
	service_mock "github.com/nt0xa/sonar/internal/service/mock"
)

func connect(t *testing.T, svc service.Service) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	st, ct := mcp.NewInMemoryTransports()
	ss, err := mcpserver.New(svc, "test").Connect(ctx, st, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ss.Close() })

	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil).Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })

	return cs
}

func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	return res
}

func text(res *mcp.CallToolResult) string {
	return res.Content[0].(*mcp.TextContent).Text
}

func TestToolsCoverService(t *testing.T) {
	cs := connect(t, &service_mock.ServerService{})

	res, err := cs.ListTools(context.Background(), nil)
	require.NoError(t, err)

	// Admin-only operations (users and audit records, 5 methods) are not exposed.
	assert.Len(t, res.Tools, reflect.TypeFor[service.Service]().NumMethod()-5)
}

func TestHTTPRoutesCreateDefaults(t *testing.T) {
	svc := &service_mock.ServerService{}
	svc.On("HTTPRoutesCreate", mock.Anything, service.HTTPRoutesCreateInput{
		PayloadName: "test",
		Method:      service.HTTPMethodGET,
		Path:        "/",
		Code:        200,
		Body:        "aGVsbG8=", // "hello"
	}).Return(&service.HTTPRoutesCreateOutput{Index: 1, Path: "/"}, nil)

	res := call(t, connect(t, svc), "http_routes_create", map[string]any{
		"payloadName": "test",
		"body":        "hello",
	})

	assert.False(t, res.IsError, text(res))
	assert.Contains(t, text(res), `"index":1`)
	svc.AssertExpectations(t)
}

func TestPayloadsCreateNoNotifyProtocols(t *testing.T) {
	svc := &service_mock.ServerService{}
	svc.On("PayloadsCreate", mock.Anything, service.PayloadsCreateInput{
		Name:            "test",
		NotifyProtocols: []service.ProtoCategory{},
	}).Return(&service.PayloadsCreateOutput{Name: "test"}, nil)

	res := call(t, connect(t, svc), "payloads_create", map[string]any{"name": "test"})

	assert.False(t, res.IsError, text(res))
	svc.AssertExpectations(t)
}

func TestServiceError(t *testing.T) {
	svc := &service_mock.ServerService{}
	svc.On("PayloadsDelete", mock.Anything, service.PayloadsDeleteInput{Name: "test"}).
		Return(nil, service.NotFoundf("payload %q not found", "test"))

	res := call(t, connect(t, svc), "payloads_delete", map[string]any{"name": "test"})

	assert.True(t, res.IsError)
	assert.Contains(t, text(res), `payload "test" not found`)
}

func TestMissingRequiredArgument(t *testing.T) {
	res := call(t, connect(t, &service_mock.ServerService{}), "events_get", map[string]any{"payloadName": "test"})

	assert.True(t, res.IsError)
	assert.Contains(t, text(res), "index")
}
