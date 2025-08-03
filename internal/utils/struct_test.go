package utils_test

import (
	"testing"

	"github.com/nt0xa/sonar/internal/utils"
	"github.com/stretchr/testify/assert"
)

type Inner struct {
	B string
	C int
}

type Outer struct {
	A string
	D *Inner
	E []int
	F map[string]string
	G Inner
	H []Inner
}

// Test for StructToMap
func TestStructToMap(t *testing.T) {
	tests := []struct {
		name   string
		input  Outer
		expect map[string]any
	}{
		{
			name: "All fields non-empty",
			input: Outer{
				A: "hello",
				D: &Inner{B: "test", C: 42},
				E: []int{1, 2, 3},
				F: map[string]string{"foo": "bar"},
				G: Inner{B: "test", C: 42},
			},
			expect: map[string]any{
				"A": "hello",
				"D": map[string]any{
					"B": "test",
					"C": 42,
				},
				"E": []any{1, 2, 3},
				"F": map[string]any{"foo": "bar"},
				"G": map[string]any{"B": "test", "C": 42},
			},
		},
		{
			name: "Empty nested struct, nil pointer, empty slice/map",
			input: Outer{
				A: "",
				D: &Inner{B: "", C: 0},
				E: nil,
				F: nil,
				G: Inner{},
			},
			expect: map[string]any{},
		},
		{
			name: "Some fields empty",
			input: Outer{
				A: "world",
				E: []int{},
			},
			expect: map[string]any{
				"A": "world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.StructToMap(tt.input)
			assert.EqualValues(t, tt.expect, got)
		})
	}
}
