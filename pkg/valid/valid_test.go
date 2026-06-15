package valid_test

import (
	"errors"
	"math"
	"regexp"
	"testing"

	"github.com/nt0xa/sonar/pkg/valid"
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
	got := valid.Validate(
		valid.String("name", "abc", valid.Required, valid.MinLength(2)),
		valid.String("color", colorRed, valid.In(colorValues()...)),
		valid.Slice("tags", []string{"a", "b"}, valid.NotEmpty, valid.Each(notFoo)),
		valid.Slice("ports", []int{80, 443}, valid.Each(valid.Min(1), valid.Max(65535))),
		valid.Number("count", 10, valid.Required, valid.Min(1), valid.Max(100)),
		valid.OptionalString("path", &path, valid.Match(pathRe, "bad path")),
		valid.OptionalString("missing", (*string)(nil), valid.Required),
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
		field   valid.Validatable
		wantKey string
		wantMsg string
	}{
		{"required string", valid.String("name", "", valid.Required), "name", "is required"},
		{"min length", valid.String("name", "a", valid.MinLength(3)), "name", "must be at least 3 characters"},
		{"max length", valid.String("name", "abcd", valid.MaxLength(3)), "name", "must be at most 3 characters"},
		{"not blank", valid.String("name", "  ", valid.NotBlank), "name", "must not be blank"},
		{"in", valid.String("color", color("green"), valid.In(colorValues()...)), "color", "must be one of: red, blue"},
		{"match", valid.String("path", "x", valid.Match(pathRe, "bad path")), "path", "bad path"},
		{"custom func", valid.String("x", "foo", notFoo), "x", "must not be foo"},
		{"required number", valid.Number("index", 0, valid.Required), "index", "is required"},
		{"min", valid.Number("n", 5, valid.Min(10)), "n", "must be >= 10"},
		{"max", valid.Number("n", 200, valid.Max(100)), "n", "must be <= 100"},
		{"not empty slice", valid.Slice("tags", []string(nil), valid.NotEmpty), "tags", "is required"},
		{"each element", valid.Slice("tags", []string{"ok", "foo"}, valid.Each(notFoo)), "tags", "element #1: must not be foo"},
		{"each in", valid.Slice("colors", []color{colorRed, "x"}, valid.Each(valid.In(colorValues()...))), "colors", "element #1: must be one of: red, blue"},
		{"each min", valid.Slice("ports", []int{0, 5}, valid.Each(valid.Min(1))), "ports", "element #0: must be >= 1"},
		{"optional ptr", valid.OptionalString("path", new("x"), valid.Match(pathRe, "bad path")), "path", "bad path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valid.Validate(tt.field)
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
	if got := valid.Validate(valid.Number("u", big, valid.Required, valid.Min(uint64(1)))); !got.Ok() {
		t.Fatalf("large uint64 should pass Min(1), got %v", got)
	}
	if got := valid.Validate(valid.Number("f", 1.5, valid.Max(1.0))); got["f"] != "must be <= 1" {
		t.Fatalf("float max: got %v", got)
	}
}

// inner is a nested Validatable.
type inner struct {
	Host string
}

func (i inner) Validate() valid.Problems {
	return valid.Validate(valid.String("host", i.Host, valid.Required))
}

// Three levels of nesting to prove keys compose as a.b.c.
type lvl3 struct{ C string }

func (l lvl3) Validate() valid.Problems {
	return valid.Validate(valid.String("c", l.C, valid.Required))
}

type lvl2 struct{ B lvl3 }

func (l lvl2) Validate() valid.Problems { return valid.Validate(valid.Struct("b", l.B)) }

func TestStruct_Nested(t *testing.T) {
	// Single nested struct: child key is namespaced under the field name.
	got := valid.Validate(valid.Struct("inner", inner{Host: ""}))
	if got["inner.host"] != "is required" {
		t.Fatalf("got %v, want inner.host=is required", got)
	}

	// Valid nested struct contributes nothing.
	got = valid.Validate(valid.Struct("inner", inner{Host: "example.com"}))
	if !got.Ok() {
		t.Fatalf("expected ok, got %v", got)
	}

	// Deep nesting composes: a.b.c.
	got = valid.Validate(valid.Struct("a", lvl2{}))
	if got["a.b.c"] != "is required" {
		t.Fatalf("got %v, want a.b.c=is required", got)
	}

	// A nil value is skipped (optional nested struct).
	if got := valid.Validate(valid.Struct("inner", nil)); !got.Ok() {
		t.Fatalf("nil value should be ok, got %v", got)
	}
}

func TestStruct_MixedLeafAndNested(t *testing.T) {
	got := valid.Validate(
		valid.String("name", "", valid.Required),
		valid.Struct("inner", inner{Host: ""}),
	)
	if got["name"] != "is required" || got["inner.host"] != "is required" {
		t.Fatalf("got %v, want both name and inner.host", got)
	}
}

func TestStruct_FirstProblemPerFieldWins(t *testing.T) {
	// Within a field, the first failing rule wins.
	got := valid.Validate(valid.String("name", "", valid.Required, valid.MinLength(5)))
	if got["name"] != "is required" {
		t.Fatalf("got %q, want first rule's message", got["name"])
	}

	// Across fields, the first occurrence of a name is kept.
	got = valid.Validate(
		valid.String("name", "", valid.Required),
		valid.String("name", "a", valid.MinLength(5)),
	)
	if got["name"] != "is required" {
		t.Fatalf("got %q, want first field's message", got["name"])
	}
}

func TestFormatRules(t *testing.T) {
	tests := []struct {
		name string
		rule valid.Rule[string]
		ok   []string
		bad  []string
	}{
		{"IP", valid.IP, []string{"127.0.0.1", "::1", "2001:db8::1"}, []string{"", "256.0.0.1", "example.com"}},
		{"Domain", valid.Domain, []string{"example.com", "a.b.example.org"}, []string{"", "example", "-bad.com", "http://x.com"}},
		{"URL", valid.URL, []string{"https://example.com", "http://x:8080/p"}, []string{"", "example.com", "/just/path"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, s := range tt.ok {
				if err := tt.rule(s); err != nil {
					t.Errorf("%q: expected ok, got %v", s, err)
				}
			}
			for _, s := range tt.bad {
				if err := tt.rule(s); err == nil {
					t.Errorf("%q: expected error, got nil", s)
				}
			}
		})
	}
}

func TestProblems_Error(t *testing.T) {
	p := valid.Validate(
		valid.String("name", "", valid.Required),
		valid.String("ip", "bad", valid.IP),
	)
	// Keys are sorted: ip before name.
	if got, want := p.Error(), "ip: must be a valid IP address; name: is required"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
