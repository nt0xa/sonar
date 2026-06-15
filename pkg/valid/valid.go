// Package valid provides declarative, type-safe struct validation. A field is a
// name, a value, and a list of rules; a rule is just `func(T) error`, so any
// custom validator is a first-class rule with no wrapper:
//
//	func (in CreateInput) Validate() valid.Problems {
//		return valid.Validate(
//			valid.String("name", in.Name, valid.Required, valid.MinLength(3)),
//			valid.String("host", in.Host, valid.Required, isSubdomain), // isSubdomain: func(string) error
//			valid.Number("ttl", in.TTL, valid.Required, valid.Max(100)),
//			valid.Slice("values", in.Values, valid.NotEmpty, valid.Each(isIPv4)),
//		)
//	}
//
// Type safety comes from the per-type constructors: String only accepts
// Rule[~string], Number only Rule[number], Slice only Rule[[]T] — so a string
// rule cannot be attached to a number field, and Each([]T)'s element rules are
// keyed to the element type.
package valid

import (
	"errors"
	"sort"
	"strings"
)

// Validatable is anything that validates itself, returning a flat map of keyed
// problems (nil when ok). Both the leaf fields built by String/Number/Slice and
// domain types (e.g. service Input types exposing Validate() Problems) satisfy
// it, so a Validatable can be a single field in a Validate(...) call or a nested
// struct folded in via Struct, which namespaces its keys.
type Validatable interface {
	// Validate returns problems keyed relative to this value: a leaf returns
	// {name: msg}, a nested struct returns {name.child: msg, ...}; nil when ok.
	Validate() Problems
}

// Rule validates a value of type T, returning an error whose message becomes the
// field's problem, or nil when the value is ok.
type Rule[T any] func(T) error

// Problems maps a json-style field name to its validation problem. nil means no
// problems. It is a defined type so it reads as a domain type at call sites and
// leaves room to grow, while still marshaling like its underlying map and being
// assignable to a plain map[string]string.
type Problems map[string]string

// Ok reports whether there are no problems.
func (p Problems) Ok() bool { return len(p) == 0 }

// Error renders problems as "key: msg; key2: msg2" with keys sorted, letting a
// non-empty Problems be returned where an error is expected. Guard with Ok() to
// return a nil error when there are no problems.
func (p Problems) Error() string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + ": " + p[k]
	}
	return strings.Join(parts, "; ")
}

// Validate runs the given fields and merges their problems into a single flat
// map of field name -> problem. Within a field the first failing rule wins; if
// two fields produce the same key the first is kept. It returns nil when
// everything is valid.
func Validate(fields ...Validatable) Problems {
	problems := Problems{}

	for _, field := range fields {
		for name, msg := range field.Validate() {
			if _, exists := problems[name]; !exists {
				problems[name] = msg
			}
		}
	}

	if len(problems) == 0 {
		return nil
	}

	return problems
}

// prefixed returns child with each key prefixed by "prefix.", or nil when child
// has no problems. It is how nested validators namespace their keys.
func prefixed(prefix string, child Problems) Problems {
	if child.Ok() {
		return nil
	}
	out := make(Problems, len(child))
	for k, msg := range child {
		out[prefix+"."+k] = msg
	}
	return out
}

// field is the shared implementation for all typed constructors: a name, a
// value, and an ordered list of rules.
type field[T any] struct {
	name  string
	value T
	rules []Rule[T]
}

func newField[T any](name string, value T, rules []Rule[T]) field[T] {
	return field[T]{name: name, value: value, rules: rules}
}

func (f field[T]) Validate() Problems {
	for _, rule := range f.rules {
		if err := rule(f.value); err != nil {
			return Problems{f.name: err.Error()}
		}
	}
	return nil
}

// Required asserts a comparable value is not its zero value (empty string, zero
// number, zero-value enum). For slices use NotEmpty.
func Required[T comparable](v T) error {
	var zero T
	if v == zero {
		return errors.New("is required")
	}
	return nil
}
