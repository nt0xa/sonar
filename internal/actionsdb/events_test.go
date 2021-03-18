package actionsdb_test

import (
	"context"
	"testing"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/actionsdb"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventsList_Success(t *testing.T) {
	u, err := db.UsersGetByID(1)
	require.NoError(t, err)

	ctx := actionsdb.SetUser(context.Background(), u)

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
				Count:       10,
			},
			9,
			1, 9,
		},
		{
			"count",
			actions.EventsListParams{
				PayloadName: "payload1",
				Count:       5,
			},
			5,
			1, 5,
		},
		{
			"after",
			actions.EventsListParams{
				PayloadName: "payload1",
				Count:       5,
				After:       4,
			},
			5,
			5, 9,
		},
		{
			"before",
			actions.EventsListParams{
				PayloadName: "payload1",
				Count:       5,
				Before:      9,
			},
			5,
			4, 8,
		},
		{
			"reverse",
			actions.EventsListParams{
				PayloadName: "payload1",
				Count:       5,
				Before:      9,
				Reverse:     true,
			},
			5,
			8, 4,
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

			ctx := context.Background()

			if tt.userID != 0 {
				u, err := db.UsersGetByID(1)
				require.NoError(t, err)

				ctx = actionsdb.SetUser(context.Background(), u)
			}

			_, err := acts.EventsList(ctx, tt.p)
			assert.Error(t, err)
			assert.IsType(t, tt.err, err)
		})
	}
}
