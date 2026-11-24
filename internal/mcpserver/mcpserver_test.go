package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/mcpserver"
	"github.com/nt0xa/sonar/internal/service"
	service_mock "github.com/nt0xa/sonar/internal/service/mock"
)

func callTool(t *testing.T, svc service.Service, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	ctx := context.Background()

	st, ct := mcp.NewInMemoryTransports()
	ss, err := mcpserver.New(svc, "test").Connect(ctx, st, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ss.Close() })

	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil).Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)

	return res
}

func TestCallTool(t *testing.T) {
	svc := &service_mock.ServerService{}
	svc.On("PayloadsDelete", mock.Anything, service.PayloadsDeleteInput{Name: "test"}).
		Return(&service.PayloadsDeleteOutput{Name: "test"}, nil)

	res := callTool(t, svc, "payloads_delete", map[string]any{"name": "test"})

	assert.False(t, res.IsError)
	assert.Contains(t, res.Content[0].(*mcp.TextContent).Text, `"name":"test"`)
	svc.AssertExpectations(t)
}

func TestCallToolError(t *testing.T) {
	svc := &service_mock.ServerService{}
	svc.On("PayloadsDelete", mock.Anything, service.PayloadsDeleteInput{Name: "test"}).
		Return(nil, service.NotFoundf("payload %q not found", "test"))

	res := callTool(t, svc, "payloads_delete", map[string]any{"name": "test"})

	assert.True(t, res.IsError)
	assert.Contains(t, res.Content[0].(*mcp.TextContent).Text, `payload "test" not found`)
}
