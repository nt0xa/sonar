package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nt0xa/sonar/pkg/valid"
)

func TestEventsListInput_Validate(t *testing.T) {
	tests := []struct {
		name string
		in   EventsListInput
		want valid.Problems
	}{
		{"defaults", EventsListInput{PayloadName: "p"}, nil},
		{"max count", EventsListInput{PayloadName: "p", Count: 100}, nil},
		{"reverse", EventsListInput{PayloadName: "p", Reverse: true}, nil},
		{"offset", EventsListInput{PayloadName: "p", Offset: 5}, nil},
		{"after zero", EventsListInput{PayloadName: "p", After: new(int64(0))}, nil},
		{"before", EventsListInput{PayloadName: "p", Before: new(int64(5))}, nil},
		{"count too big", EventsListInput{PayloadName: "p", Count: 101},
			valid.Problems{"count": "must be <= 100"}},
		{"after negative", EventsListInput{PayloadName: "p", After: new(int64(-1))},
			valid.Problems{"after": "must be >= 0"}},
		{"before zero", EventsListInput{PayloadName: "p", Before: new(int64(0))},
			valid.Problems{"before": "must be >= 1"}},
		{"after and before", EventsListInput{PayloadName: "p", After: new(int64(1)), Before: new(int64(5))},
			valid.Problems{"before": "cannot be used with after"}},
		{"offset and after", EventsListInput{PayloadName: "p", Offset: 1, After: new(int64(1))},
			valid.Problems{"offset": "cannot be used with after, before or reverse"}},
		{"offset and before", EventsListInput{PayloadName: "p", Offset: 1, Before: new(int64(5))},
			valid.Problems{"offset": "cannot be used with after, before or reverse"}},
		{"offset and reverse", EventsListInput{PayloadName: "p", Offset: 1, Reverse: true},
			valid.Problems{"offset": "cannot be used with after, before or reverse"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.Validate())
		})
	}
}
