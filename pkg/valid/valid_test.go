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

func notFoo(s string) error {
	if s == "foo" {
		return errors.New("must not be foo")
	}
	return nil
}

func TestStruct_Valid(t *testing.T) {
	path := "/ok"
	got := v.Struct(
		v.String("name", "abc").Required().MinLength(2),
		v.String("color", colorRed).In(colorValues()...),
		v.StringSlice("tags", []string{"a", "b"}).Required().Each().Custom(notFoo),
		v.NumberSlice("ports", []int{80, 443}).MaxItems(100).Each().Min(1).Max(65535),
		v.Number("count", 10).Required().Min(1).Max(100),
		v.OptionalString("path", &path).Match(pathRe, "bad path"),
		v.OptionalString("missing", (*string)(nil)).In("x"),
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
		{"required string", v.String("name", "").Required(), "name", "is required"},
		{"min length", v.String("name", "a").MinLength(3), "name", "must be at least 3 characters"},
		{"max length", v.String("name", "abcd").MaxLength(3), "name", "must be at most 3 characters"},
		{"not blank", v.String("name", "  ").NotBlank(), "name", "must not be blank"},
		{"in", v.String("color", color("green")).In(colorValues()...), "color", "must be one of: red, blue"},
		{"match", v.String("path", "x").Match(pathRe, "bad path"), "path", "bad path"},
		{"custom", v.String("x", "foo").Custom(notFoo), "x", "must not be foo"},
		{"required number", v.Number("index", 0).Required(), "index", "is required"},
		{"min", v.Number("n", 5).Min(10), "n", "must be >= 10"},
		{"max", v.Number("n", 200).Max(100), "n", "must be <= 100"},
		{"required slice", v.StringSlice("tags", []string(nil)).Required(), "tags", "is required"},
		{"min items", v.StringSlice("tags", []string{"a"}).MinItems(2), "tags", "must contain at least 2 items"},
		{"max items", v.StringSlice("tags", []string{"a", "b"}).MaxItems(1), "tags", "must contain no more than 1 items"},
		{"each element", v.StringSlice("tags", []string{"ok", "foo"}).Each().Custom(notFoo), "tags", "element #1: must not be foo"},
		{"each in", v.StringSlice("colors", []color{colorRed, "x"}).Each().In(colorValues()...), "colors", "element #1: must be one of: red, blue"},
		{"number slice min", v.NumberSlice("ports", []int{0, 5}).Each().Min(1), "ports", "element #0: must be >= 1"},
		{"slice len before each", v.NumberSlice("ports", []int{1, 2, 3}).MaxItems(2).Each().Min(1), "ports", "must contain no more than 2 items"},
		{"optional ptr", v.OptionalString("path", ptr("x")).Match(pathRe, "bad path"), "path", "bad path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.Struct(tt.field)
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
	if got := v.Struct(v.Number("u", big).Required().Min(uint64(1))); !got.Ok() {
		t.Fatalf("large uint64 should pass Min(1), got %v", got)
	}
	if got := v.Struct(v.Number("f", 1.5).Max(1.0)); got["f"] != "must be <= 1" {
		t.Fatalf("float max: got %v", got)
	}
}

func TestStruct_FirstProblemPerFieldWins(t *testing.T) {
	// Within a field, the first failing rule wins.
	got := v.Struct(v.String("name", "").Required().MinLength(5))
	if got["name"] != "is required" {
		t.Fatalf("got %q, want first rule's message", got["name"])
	}

	// Across fields, the first occurrence of a name is kept.
	got = v.Struct(
		v.String("name", "").Required(),
		v.String("name", "a").MinLength(5),
	)
	if got["name"] != "is required" {
		t.Fatalf("got %q, want first field's message", got["name"])
	}
}

func ptr[T any](v T) *T { return &v }
