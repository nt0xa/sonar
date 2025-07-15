package database_test

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
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

	err := db.EventsCreate(t.Context(), o)
	assert.NoError(t, err)
	assert.NotZero(t, o.ID)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)
}

func TestEventsGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsGetByID(t.Context(), 1)
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

	o, err := db.EventsGetByID(t.Context(), 1337)
	assert.Error(t, err)
	assert.Nil(t, o)
	assert.Error(t, err, sql.ErrNoRows.Error())
}

func TestEventsListByPayloadID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	// Default
	l, err := db.EventsListByPayloadID(t.Context(), 1)
	assert.NoError(t, err)
	require.Len(t, l, 10)
	assert.EqualValues(t, l[0].ID, 11)
	assert.EqualValues(t, l[len(l)-1].ID, 1)

	l, err = db.EventsListByPayloadID(t.Context(), 1,
		database.EventsPagination(database.Page{
			Count: 3,
		}),
	)

	// Count
	assert.NoError(t, err)
	require.Len(t, l, 3)
	assert.EqualValues(t, l[0].ID, 11)
	assert.EqualValues(t, l[len(l)-1].ID, 8)

	// Before
	l, err = db.EventsListByPayloadID(t.Context(), 1,
		database.EventsPagination(database.Page{
			Count:  5,
			Before: 7,
		}),
	)
	assert.NoError(t, err)
	require.Len(t, l, 5)
	assert.EqualValues(t, l[0].ID, 6)
	assert.EqualValues(t, l[len(l)-1].ID, 2)

	// After
	l, err = db.EventsListByPayloadID(t.Context(), 1,
		database.EventsPagination(database.Page{
			Count: 5,
			After: 7,
		}),
	)
	assert.NoError(t, err)
	require.Len(t, l, 3)
	assert.EqualValues(t, l[0].ID, 11)
	assert.EqualValues(t, l[len(l)-1].ID, 8)

	// Reverse
	l, err = db.EventsListByPayloadID(t.Context(), 1,
		database.EventsPagination(database.Page{
			Count: 5,
			After: 7,
		}),
		database.EventsReverse(true),
	)
	assert.NoError(t, err)
	require.Len(t, l, 3)
	assert.EqualValues(t, l[0].ID, 8)
	assert.EqualValues(t, l[len(l)-1].ID, 11)
}

func TestEventsGetByPayloadAndIndex_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsGetByPayloadAndIndex(t.Context(), 1, 1)
	assert.NoError(t, err)
	assert.EqualValues(t, o.ID, 1)

	o, err = db.EventsGetByPayloadAndIndex(t.Context(), 1, 1337)
	assert.Error(t, err)
}

func TestEventsRace(t *testing.T) {
	setup(t)
	defer teardown(t)

	var wg sync.WaitGroup
	count := 10

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

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

			err := db.EventsCreate(t.Context(), o)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}
