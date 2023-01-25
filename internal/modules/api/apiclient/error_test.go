package apiclient_test

import (
	"testing"

	"github.com/russtone/sonar/internal/modules/api/apiclient"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {

	e := apiclient.APIError{Msg: "message", Det: "details", Errs: map[string]interface{}{
		"field": "error",
	}}

	assert.Equal(t, "message", e.Message())
	assert.Equal(t, "details", e.Details())
	assert.Contains(t, e.Error(), "field")
}
