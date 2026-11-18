package database_test

import (
	"math"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/pkg/dnsx"
)

func TestEventsCreate_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsCreate(t.Context(), database.EventsCreateParams{
		PayloadID: 1,
		UUID:      uuid.New(),
		Protocol:  "dns",
		Data:      [][]byte{{1, 3, 5}, {2, 4}},
		Meta: database.EventsMeta{
			DNS: &dnsx.Meta{
				Question: dnsx.Question{Name: "test.example.com", Type: "A"},
			},
		},
		ReceivedAt: time.Now(),
		RemoteAddr: "127.0.0.1:1337",
	})
	assert.NoError(t, err)
	assert.NotZero(t, o.ID)
	assert.EqualValues(t, 11, o.Index)
	assert.WithinDuration(t, time.Now(), o.CreatedAt, 5*time.Second)
}

func TestEventsCreate_IndexNotReused(t *testing.T) {
	setup(t)
	defer teardown(t)

	_, err := db.DB().ExecContext(t.Context(), "DELETE FROM events WHERE payload_id = 1 AND index = 10")
	require.NoError(t, err)

	o, err := db.EventsCreate(t.Context(), database.EventsCreateParams{
		PayloadID:  1,
		UUID:       uuid.New(),
		Protocol:   "dns",
		Data:       [][]byte{},
		ReceivedAt: time.Now(),
		RemoteAddr: "127.0.0.1:1337",
	})
	require.NoError(t, err)
	assert.EqualValues(t, 11, o.Index)
}

func TestEventsGetByID_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsGetByID(t.Context(), 1)
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.EqualValues(t, 1, o.PayloadID)
	assert.Equal(t, [][]byte{[]byte("read"), []byte("written")}, o.Data)
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

	tests := []struct {
		name   string
		params database.EventsListByPayloadIDParams
		want   []int64 // indexes
	}{
		{"all", database.EventsListByPayloadIDParams{PayloadID: 1, Before: math.MaxInt64, PageLimit: 100},
			[]int64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{"limit", database.EventsListByPayloadIDParams{PayloadID: 1, Before: math.MaxInt64, PageLimit: 3},
			[]int64{10, 9, 8}},
		{"offset", database.EventsListByPayloadIDParams{PayloadID: 1, Before: math.MaxInt64, PageLimit: 5, PageOffset: 5},
			[]int64{5, 4, 3, 2, 1}},
		{"before", database.EventsListByPayloadIDParams{PayloadID: 1, Before: 5, PageLimit: 3},
			[]int64{4, 3, 2}},
		{"before first", database.EventsListByPayloadIDParams{PayloadID: 1, Before: 1, PageLimit: 3},
			[]int64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := db.EventsListByPayloadID(t.Context(), tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, indexes(l))
		})
	}
}

func TestEventsListByPayloadIDAfter_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	tests := []struct {
		name   string
		params database.EventsListByPayloadIDAfterParams
		want   []int64 // indexes
	}{
		{"from start", database.EventsListByPayloadIDAfterParams{PayloadID: 1, After: 0, PageLimit: 3},
			[]int64{1, 2, 3}},
		{"after", database.EventsListByPayloadIDAfterParams{PayloadID: 1, After: 7, PageLimit: 5},
			[]int64{8, 9, 10}},
		{"after last", database.EventsListByPayloadIDAfterParams{PayloadID: 1, After: 10, PageLimit: 5},
			[]int64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := db.EventsListByPayloadIDAfter(t.Context(), tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, indexes(l))
		})
	}
}

func TestEventsGetByPayloadAndIndex_Success(t *testing.T) {
	setup(t)
	defer teardown(t)

	o, err := db.EventsGetByPayloadAndIndex(t.Context(), 1, 1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, o.ID)

	o, err = db.EventsGetByPayloadAndIndex(t.Context(), 1, 10)
	assert.NoError(t, err)
	assert.EqualValues(t, 11, o.ID)

	_, err = db.EventsGetByPayloadAndIndex(t.Context(), 1, 1337)
	assert.Error(t, err)
}

func TestEventsRace(t *testing.T) {
	setup(t)
	defer teardown(t)

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		got []int64
	)
	count := 10

	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()

			o, err := db.EventsCreate(t.Context(), database.EventsCreateParams{
				PayloadID: 1,
				UUID:      uuid.New(),
				Protocol:  "dns",
				Data:      [][]byte{{1, 3, 5}, {2, 4}},
				Meta: database.EventsMeta{
					DNS: &dnsx.Meta{
						Question: dnsx.Question{Name: "test.example.com", Type: "A"},
					},
				},
				ReceivedAt: time.Now(),
				RemoteAddr: "127.0.0.1:1337",
			})
			if assert.NoError(t, err) {
				mu.Lock()
				got = append(got, o.Index)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	slices.Sort(got)
	assert.Equal(t, []int64{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, got)
}

func indexes(l []*database.Event) []int64 {
	out := make([]int64, len(l))
	for i, e := range l {
		out[i] = e.Index
	}
	return out
}
