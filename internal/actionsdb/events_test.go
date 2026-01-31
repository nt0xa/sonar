package actionsdb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func TestEventsList_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name        string
		p           actions.EventsListParams
		count       int
		first, last int
	}{
		{
			"all",
			actions.EventsListParams{
				PayloadName: "payload1",
				Limit:       10,
			},
			10,
			10, 1,
		},
		{
			"count",
			actions.EventsListParams{
				PayloadName: "payload1",
				Limit:       5,
			},
			5,
			10, 6,
		},
		{
			"offset",
			actions.EventsListParams{
				PayloadName: "payload1",
				Limit:       5,
				Offset:      5,
			},
			5,
			5, 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			r, err := acts.EventsList(ctx, tt.p)
			assert.NoError(t, err)
			assert.Len(t, r, tt.count)

			assert.EqualValues(t, tt.first, r[0].Index)
			assert.EqualValues(t, tt.last, r[len(r)-1].Index)
		})
	}
}

func TestEventsList_Error(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
		p      actions.EventsListParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
			actions.EventsListParams{
				PayloadName: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			1,
			actions.EventsListParams{
				PayloadName: "",
			},
			&errors.ValidationError{},
		},
		{
			"non existent payload",
			1,
			actions.EventsListParams{
				PayloadName: "non-exist",
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := t.Context()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(t.Context(), 1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(t.Context(), u)
			}

			_, err := acts.EventsList(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}

func TestEventsGet_Success(t *testing.T) {
	u, err := db.UsersGetByID(t.Context(), 1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(t.Context(), u)

	tests := []struct {
		name     string
		p        actions.EventsGetParams
		protocol string
	}{
		{
			"all",
			actions.EventsGetParams{
				PayloadName: "payload1",
				Index:       3,
			},
			"http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			e, err := acts.EventsGet(ctx, tt.p)
			require.NoError(t, err)

			assert.EqualValues(t, tt.protocol, e.Protocol)
		})
	}
}

func TestEventsGet_Error(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
		p      actions.EventsGetParams
		err    errors.Error
	}{
		{
			"no user in ctx",
			0,
			actions.EventsGetParams{
				PayloadName: "test",
			},
			&errors.InternalError{},
		},
		{
			"empty payload name",
			1,
			actions.EventsGetParams{
				PayloadName: "",
			},
			&errors.ValidationError{},
		},
		{
			"invalid index",
			1,
			actions.EventsGetParams{
				PayloadName: "payload1",
			},
			&errors.ValidationError{},
		},
		{
			"non existent payload",
			1,
			actions.EventsGetParams{
				PayloadName: "non-exist",
				Index:       1,
			},
			&errors.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			defer teardown(t)

			ctx := t.Context()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(t.Context(), 1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(t.Context(), u)
			}

			_, err := acts.EventsGet(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
