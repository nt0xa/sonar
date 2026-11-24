package dbsvc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/service"
)

func TestHTTPRoutesCreate_NilHeaders(t *testing.T) {
	setup(t)

	ctx := service.WithCaller(t.Context(), service.Caller{UserID: 1})

	out, err := svc.HTTPRoutesCreate(ctx, service.HTTPRoutesCreateInput{
		PayloadName: "payload1",
		Method:      service.HTTPMethodGET,
		Path:        "/nil-headers",
		Code:        200,
	})
	require.NoError(t, err)
	assert.Equal(t, map[string][]string{}, out.Headers)
}
