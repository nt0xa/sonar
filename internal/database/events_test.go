package database_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bi-zone/sonar/internal/models"
)

func TestEventsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o := &models.Event{
		PayloadID: 1,
		Protocol:  models.ProtoDNS,
		R:         []byte{1, 3, 5},
		W:         []byte{2, 4},
		RW:        []byte{1, 2, 3, 4, 5},
		Meta: models.Meta{
			"key": "value",
		},
		ReceivedAt: time.Now(),
		RemoteAddr: "127.0.0.1:1337",
	}

	err := db.EventsCreate(o)
	assert.NoError(t, err)
	assert.NotZero(t, o.ID)
	assert.WithinDuration(t, time.Now().UTC(), o.CreatedAt, 5*time.Second)
}

func TestEventsGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsGetByID(1)
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.EqualValues(t, 1, o.PayloadID)
	assert.Equal(t, []byte("read"), o.R)
	assert.Equal(t, []byte("written"), o.W)
	assert.Equal(t, []byte("read-and-written"), o.RW)
	assert.Equal(t, models.Meta{"key": "value"}, o.Meta)
	assert.Equal(t, "127.0.0.1:1337", o.RemoteAddr)
}

func TestEventsGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsGetByID(1337)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.Error(t, err, sql.ErrNoRows.Error())
}

func TestEventsFindByPayloadID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	l, err := db.EventsFindByPayloadID(1)
	assert.NoError(t, err)
	require.Len(t, l, 2)

	assert.EqualValues(t, l[0].ID, 2)
	assert.EqualValues(t, l[1].ID, 1)
}
