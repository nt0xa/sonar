package utils_test

import (
	"testing"

	"github.com/nt0xa/sonar/internal/utils"
	"github.com/stretchr/testify/assert"
)

type Inner struct {
	B string `audit:"b"`
	C int    `audit:"c"`
}

type Outer struct {
	A string            `audit:"a"`
	D *Inner            `audit:"d"`
	E []int             `audit:"e"`
	F map[string]string `audit:"f"`
	G Inner             `audit:"g"`
	H []Inner
	O int    `audit:"o,omitempty"`
	P string `audit:"p,omitempty"`
	Q bool   `audit:"q,omitempty"`
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
				O: 7,
				P: "x",
				Q: true,
			},
			expect: map[string]any{
				"a": "hello",
				"d": Inner{B: "test", C: 42},
				"e": []int{1, 2, 3},
				"f": map[string]string{"foo": "bar"},
				"g": Inner{B: "test", C: 42},
				"o": 7,
				"p": "x",
				"q": true,
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
			expect: map[string]any{
				"a": "",
				"d": Inner{},
				"e": []int(nil),
				"f": map[string]string(nil),
				"g": Inner{},
			},
		},
		{
			name: "Some fields empty",
			input: Outer{
				A: "world",
				E: []int{},
			},
			expect: map[string]any{
				"a": "world",
				"e": []int{},
				"f": map[string]string(nil),
				"g": Inner{},
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
