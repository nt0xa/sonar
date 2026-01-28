package database2_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database2"
	"github.com/nt0xa/sonar/pkg/dnsx"
)

func TestEventsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsCreate(t.Context(), database2.EventsCreateParams{
		PayloadID: 1,
		UUID:      uuid.New(),
		Protocol:  "dns",
		R:         []byte{1, 3, 5},
		W:         []byte{2, 4},
		RW:        []byte{1, 2, 3, 4, 5},
		Meta: database2.EventsMeta{
			DNS: &dnsx.Meta{
				Question: dnsx.Question{Name: "test.example.com", Type: "A"},
			},
		},
		ReceivedAt: time.Now(),
		RemoteAddr: "127.0.0.1:1337",
	})
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
	assert.NotNil(t, o.Meta.DNS.Question)
	assert.Equal(t, "test.example.com", o.Meta.DNS.Question.Name)
	assert.Equal(t, "A", o.Meta.DNS.Question.Type)
	assert.Equal(t, "127.0.0.1:1337", o.RemoteAddr)
	assert.Equal(t, "c0b49dee-3ce9-4bd9-b111-7abd7a2f16f0", o.UUID.String())
}

func TestEventsGetByID_NotExist(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.EventsGetByID(t.Context(), 1337)
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestEventsListByPayloadID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	// All events for payload 1
	l, err := db.EventsListByPayloadID(t.Context(), database2.EventsListByPayloadIDParams{
		PayloadID: 1,
		Limit:     100,
		Offset:    0,
	})
	assert.NoError(t, err)
	require.Len(t, l, 10)
	assert.EqualValues(t, 11, l[0].ID)
	assert.EqualValues(t, 1, l[len(l)-1].ID)

	// Limit
	l, err = db.EventsListByPayloadID(t.Context(), database2.EventsListByPayloadIDParams{
		PayloadID: 1,
		Limit:     3,
		Offset:    0,
	})
	assert.NoError(t, err)
	require.Len(t, l, 3)
	assert.EqualValues(t, 11, l[0].ID)
	assert.EqualValues(t, 8, l[len(l)-1].ID)

	// Offset
	l, err = db.EventsListByPayloadID(t.Context(), database2.EventsListByPayloadIDParams{
		PayloadID: 1,
		Limit:     5,
		Offset:    5,
	})
	assert.NoError(t, err)
	require.Len(t, l, 5)
	assert.EqualValues(t, 5, l[0].ID)
	assert.EqualValues(t, 1, l[len(l)-1].ID)
}

func TestEventsGetByPayloadAndIndex_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsGetByPayloadAndIndex(t.Context(), 1, 1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, o.ID)

	_, err = db.EventsGetByPayloadAndIndex(t.Context(), 1, 1337)
	assert.Error(t, err)
}

func TestEventsRace(t *testing.T) {
	setup(t)
	defer teardown(t)

	var wg sync.WaitGroup
	count := 10

	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := db.EventsCreate(t.Context(), database2.EventsCreateParams{
				PayloadID: 1,
				UUID:      uuid.New(),
				Protocol:  "dns",
				R:         []byte{1, 3, 5},
				W:         []byte{2, 4},
				RW:        []byte{1, 2, 3, 4, 5},
				Meta: database2.EventsMeta{
					DNS: &dnsx.Meta{
						Question: dnsx.Question{Name: "test.example.com", Type: "A"},
					},
				},
				ReceivedAt: time.Now(),
				RemoteAddr: "127.0.0.1:1337",
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}
