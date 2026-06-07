package valid_test

import (
	"regexp"
	"testing"

	v "github.com/nt0xa/sonar/pkg/valid"
)

// color is a ~string type with a value list, used to verify that the generic
// field/rule constructors accept named string types.
type color string

const (
	colorRed  color = "red"
	colorBlue color = "blue"
)

func colorValues() []color { return []color{colorRed, colorBlue} }

type sample struct {
	PayloadName string
	Name        string
	Color       color
	Tags        []string
	Nums        []int
	Count       int
	ID          int64  // lowerCamel -> "id"
	APIToken    string `json:"token"` // tag override
	OptPath     *string
}

var pathRe = regexp.MustCompile("^/.*")

func (s sample) validate() v.Problems {
	return v.Struct(&s,
		v.String(&s.PayloadName, v.Required),
		v.String(&s.Name, v.Required, v.MinLength(3)),
		v.String(&s.Color, v.Required, v.In(colorValues()...)),
		v.StringSlice(&s.Tags, v.Required, v.Each(v.Required)),
		v.Slice(&s.Nums, v.MinItems(1), v.MaxItems(3)),
		v.Int(&s.Count, v.Required, v.Max(100)),
		v.Int(&s.ID, v.Required),
		v.String(&s.APIToken, v.Required),
		v.OptionalStringPtr(&s.OptPath, v.Match(pathRe, `must start with "/"`)),
	)
}

func valid() sample {
	p := "/ok"
	return sample{
		PayloadName: "p",
		Name:        "abc",
		Color:       colorBlue,
		Tags:        []string{"a"},
		Nums:        []int{1, 2},
		Count:       10,
		ID:          1,
		APIToken:    "t",
		OptPath:     &p,
	}
}

func TestStruct_Valid(t *testing.T) {
	if got := valid().validate(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestStruct_Problems(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*sample)
		wantKey string
		wantMsg string
	}{
		{"required string", func(s *sample) { s.PayloadName = "" }, "payloadName", "cannot be blank"},
		{"min length", func(s *sample) { s.Name = "ab" }, "name", "must be at least 3 characters"},
		{"in (enum)", func(s *sample) { s.Color = "green" }, "color", "must be one of: red, blue"},
		{"required slice", func(s *sample) { s.Tags = nil }, "tags", "cannot be blank"},
		{"each element", func(s *sample) { s.Tags = []string{""} }, "tags", "element #0: cannot be blank"},
		{"min items", func(s *sample) { s.Nums = nil }, "nums", "must contain at least 1 items"},
		{"max items", func(s *sample) { s.Nums = []int{1, 2, 3, 4} }, "nums", "must contain no more than 3 items"},
		{"required number", func(s *sample) { s.Count = 0 }, "count", "cannot be blank"},
		{"max", func(s *sample) { s.Count = 101 }, "count", "must be no greater than 100"},
		{"acronym field name", func(s *sample) { s.ID = 0 }, "id", "cannot be blank"},
		{"json tag override", func(s *sample) { s.APIToken = "" }, "token", "cannot be blank"},
		{"optional ptr match", func(s *sample) { *s.OptPath = "bad" }, "optPath", `must start with "/"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := valid()
			tt.mutate(&s)
			got := s.validate()
			if got[tt.wantKey] != tt.wantMsg {
				t.Fatalf("key %q: got %q, want %q (all: %v)", tt.wantKey, got[tt.wantKey], tt.wantMsg, got)
			}
		})
	}
}

func TestOptionalStringPtr_NilSkips(t *testing.T) {
	s := valid()
	s.OptPath = nil // invalid match rule must be skipped when nil
	if got := s.validate(); got != nil {
		t.Fatalf("expected nil for nil pointer, got %v", got)
	}
}

func TestOptional_SkipsWhenEmpty(t *testing.T) {
	type opt struct{ Kind string }

	o := opt{Kind: ""}
	problems := v.Struct(&o, v.String(&o.Kind, v.Optional, v.In("a", "b")))
	if problems != nil {
		t.Fatalf("optional empty should pass, got %v", problems)
	}

	o.Kind = "x"
	problems = v.Struct(&o, v.String(&o.Kind, v.Optional, v.In("a", "b")))
	if problems["kind"] == "" {
		t.Fatalf("optional non-empty invalid should fail, got %v", problems)
	}
}

func TestStruct_PanicsOnMisuse(t *testing.T) {
	type s struct{ Name string }

	cases := []struct {
		name string
		fn   func()
	}{
		{"non-pointer", func() { v.Struct(s{}) }},
		{"nil pointer", func() { v.Struct((*s)(nil)) }},
		{"pointer to non-struct", func() { n := 0; v.Struct(&n) }},
		{"foreign field pointer", func() {
			var a, b s
			b.Name = "" // make the rule fail so fieldName is consulted
			v.Struct(&a, v.String(&b.Name, v.Required))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic")
				}
			}()
			tc.fn()
		})
	}
}
