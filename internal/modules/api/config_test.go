package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bi-zone/sonar/internal/modules/api"
)

func TestConfig_Success(t *testing.T) {
	cfg := &api.Config{
		Admin: "1337",
		Port:  1337,
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}
