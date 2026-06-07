package valid_test

import (
	"errors"
	"math"
	"regexp"
	"testing"

	v "github.com/nt0xa/sonar/pkg/valid"
)

type color string

const (
	colorRed  color = "red"
	colorBlue color = "blue"
)

func colorValues() []color { return []color{colorRed, colorBlue} }

var pathRe = regexp.MustCompile("^/.*")

// notFoo is a plain func(string) error — a first-class rule with no wrapper.
func notFoo(s string) error {
	if s == "foo" {
		return errors.New("must not be foo")
	}
	return nil
}

func TestStruct_Valid(t *testing.T) {
	path := "/ok"
	got := v.Validate(
		v.String("name", "abc", v.Required, v.MinLength(2)),
		v.String("color", colorRed, v.In(colorValues()...)),
		v.Slice("tags", []string{"a", "b"}, v.NotEmpty, v.Each(notFoo)),
		v.Slice("ports", []int{80, 443}, v.Each(v.Min(1), v.Max(65535))),
		v.Number("count", 10, v.Required, v.Min(1), v.Max(100)),
		v.OptionalString("path", &path, v.Match(pathRe, "bad path")),
		v.OptionalString("missing", (*string)(nil), v.Required),
	)
	if !got.Ok() {
		t.Fatalf("expected ok, got %v", got)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestStruct_Problems(t *testing.T) {
	tests := []struct {
		name    string
		field   v.Field
		wantKey string
		wantMsg string
	}{
		{"required string", v.String("name", "", v.Required), "name", "is required"},
		{"min length", v.String("name", "a", v.MinLength(3)), "name", "must be at least 3 characters"},
		{"max length", v.String("name", "abcd", v.MaxLength(3)), "name", "must be at most 3 characters"},
		{"not blank", v.String("name", "  ", v.NotBlank), "name", "must not be blank"},
		{"in", v.String("color", color("green"), v.In(colorValues()...)), "color", "must be one of: red, blue"},
		{"match", v.String("path", "x", v.Match(pathRe, "bad path")), "path", "bad path"},
		{"custom func", v.String("x", "foo", notFoo), "x", "must not be foo"},
		{"required number", v.Number("index", 0, v.Required), "index", "is required"},
		{"min", v.Number("n", 5, v.Min(10)), "n", "must be >= 10"},
		{"max", v.Number("n", 200, v.Max(100)), "n", "must be <= 100"},
		{"not empty slice", v.Slice("tags", []string(nil), v.NotEmpty), "tags", "is required"},
		{"each element", v.Slice("tags", []string{"ok", "foo"}, v.Each(notFoo)), "tags", "element #1: must not be foo"},
		{"each in", v.Slice("colors", []color{colorRed, "x"}, v.Each(v.In(colorValues()...))), "colors", "element #1: must be one of: red, blue"},
		{"each min", v.Slice("ports", []int{0, 5}, v.Each(v.Min(1))), "ports", "element #0: must be >= 1"},
		{"optional ptr", v.OptionalString("path", new("x"), v.Match(pathRe, "bad path")), "path", "bad path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.Validate(tt.field)
			if got[tt.wantKey] != tt.wantMsg {
				t.Fatalf("key %q: got %q, want %q (all: %v)", tt.wantKey, got[tt.wantKey], tt.wantMsg, got)
			}
		})
	}
}

func TestNumber_NoCoercion(t *testing.T) {
	// A large uint64 must compare correctly (no int64 widening) and a float must
	// use float comparison.
	big := uint64(math.MaxUint64)
	if got := v.Validate(v.Number("u", big, v.Required, v.Min(uint64(1)))); !got.Ok() {
		t.Fatalf("large uint64 should pass Min(1), got %v", got)
	}
	if got := v.Validate(v.Number("f", 1.5, v.Max(1.0))); got["f"] != "must be <= 1" {
		t.Fatalf("float max: got %v", got)
	}
}

func TestStruct_FirstProblemPerFieldWins(t *testing.T) {
	// Within a field, the first failing rule wins.
	got := v.Validate(v.String("name", "", v.Required, v.MinLength(5)))
	if got["name"] != "is required" {
		t.Fatalf("got %q, want first rule's message", got["name"])
	}

	// Across fields, the first occurrence of a name is kept.
	got = v.Validate(
		v.String("name", "", v.Required),
		v.String("name", "a", v.MinLength(5)),
	)
	if got["name"] != "is required" {
		t.Fatalf("got %q, want first field's message", got["name"])
	}
}
