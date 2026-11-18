package dbsvc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/internal/service"
)

func TestEventsList(t *testing.T) {
	setup(t)

	// payload1 of user1 has events with indexes 1..10, see fixtures.
	ctx := service.WithCaller(t.Context(), service.Caller{UserID: 1})

	tests := []struct {
		name string
		in   service.EventsListInput
		want []int64
	}{
		{"latest", service.EventsListInput{Count: 3},
			[]int64{10, 9, 8}},
		{"first", service.EventsListInput{Count: 3, Reverse: true},
			[]int64{1, 2, 3}},
		{"after", service.EventsListInput{Count: 3, After: new(int64(4))},
			[]int64{7, 6, 5}},
		{"after reverse", service.EventsListInput{Count: 3, After: new(int64(4)), Reverse: true},
			[]int64{5, 6, 7}},
		{"before", service.EventsListInput{Count: 3, Before: new(int64(5))},
			[]int64{4, 3, 2}},
		{"before reverse", service.EventsListInput{Count: 3, Before: new(int64(5)), Reverse: true},
			[]int64{2, 3, 4}},
		{"offset", service.EventsListInput{Count: 3, Offset: 2},
			[]int64{8, 7, 6}},
		{"default count", service.EventsListInput{},
			[]int64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.PayloadName = "payload1"

			out, err := svc.EventsList(ctx, tt.in)
			require.NoError(t, err)

			got := make([]int64, len(out))
			for i, e := range out {
				got[i] = e.Index
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
