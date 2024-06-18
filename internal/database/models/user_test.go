package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database/models"
)

func TestUserParams(t *testing.T) {

	p := models.UserParams{
		TelegramID: 1337,
		APIToken:   "token",
	}

	value, err := p.Value()
	require.NoError(t, err)
	require.NotNil(t, value)

	err = p.Scan([]byte(`{"telegram.id": "31337", "api.token": "token2"}`))
	require.NoError(t, err)
	assert.Equal(t, p.TelegramID, int64(31337))
	assert.Equal(t, p.APIToken, "token2")
}
