package valid

import (
	"errors"
	"fmt"
	"regexp"
)

// eachBuilder is the shared state of an element builder returned by Each: the
// slice's name/value, the slice-level length rules, and the accumulated
// per-element rules. It is itself a Field.
type eachBuilder[T any] struct {
	name        string
	value       []T
	lengthRules []Rule[[]T]
	elemRules   []Rule[T]
}

func (e eachBuilder[T]) withElem(rule Rule[T]) eachBuilder[T] {
	rules := make([]Rule[T], 0, len(e.elemRules)+1)
	rules = append(rules, e.elemRules...)
	rules = append(rules, rule)

	return eachBuilder[T]{name: e.name, value: e.value, lengthRules: e.lengthRules, elemRules: rules}
}

func (e eachBuilder[T]) Validate() (string, error) {
	for _, rule := range e.lengthRules {
		if err := rule(e.value); err != nil {
			return e.name, err
		}
	}
	for i, item := range e.value {
		for _, rule := range e.elemRules {
			if err := rule(item); err != nil {
				return e.name, fmt.Errorf("element #%d: %w", i, err)
			}
		}
	}
	return e.name, nil
}

// --- string slices ---

// stringSliceField validates a slice of ~string values with slice-level length
// rules. Per-element rules are added via Each.
type stringSliceField[T ~string] struct {
	field[[]T]
}

// StringSlice validates a slice of ~string values.
func StringSlice[T ~string](name string, value []T) stringSliceField[T] {
	return stringSliceField[T]{field: newField(name, value)}
}

func (f stringSliceField[T]) withRule(rule Rule[[]T]) stringSliceField[T] {
	return stringSliceField[T]{field: f.field.withRule(rule)}
}

func (f stringSliceField[T]) Required() stringSliceField[T]      { return f.withRule(requiredSlice[T]) }
func (f stringSliceField[T]) MinItems(n int) stringSliceField[T] { return f.withRule(minItems[T](n)) }
func (f stringSliceField[T]) MaxItems(n int) stringSliceField[T] { return f.withRule(maxItems[T](n)) }

// Each transitions to a per-element builder exposing string rules.
func (f stringSliceField[T]) Each() stringSliceEach[T] {
	return stringSliceEach[T]{eachBuilder[T]{name: f.name, value: f.value, lengthRules: f.rules}}
}

type stringSliceEach[T ~string] struct {
	eachBuilder[T]
}

func (e stringSliceEach[T]) with(rule Rule[T]) stringSliceEach[T] {
	return stringSliceEach[T]{e.withElem(rule)}
}

func (e stringSliceEach[T]) Required() stringSliceEach[T]       { return e.with(requiredString[T]) }
func (e stringSliceEach[T]) NotBlank() stringSliceEach[T]       { return e.with(notBlank[T]) }
func (e stringSliceEach[T]) MinLength(n int) stringSliceEach[T] { return e.with(minLength[T](n)) }
func (e stringSliceEach[T]) MaxLength(n int) stringSliceEach[T] { return e.with(maxLength[T](n)) }
func (e stringSliceEach[T]) In(allowed ...T) stringSliceEach[T] { return e.with(OneOf(allowed...)) }

func (e stringSliceEach[T]) Match(re *regexp.Regexp, msg string) stringSliceEach[T] {
	return e.with(match[T](re, msg))
}

func (e stringSliceEach[T]) Custom(rule Rule[T]) stringSliceEach[T] { return e.with(rule) }

// --- number slices ---

// numberSliceField validates a slice of numbers with slice-level length rules.
// Per-element rules are added via Each.
type numberSliceField[T number] struct {
	field[[]T]
}

// NumberSlice validates a slice of numeric values.
func NumberSlice[T number](name string, value []T) numberSliceField[T] {
	return numberSliceField[T]{field: newField(name, value)}
}

func (f numberSliceField[T]) withRule(rule Rule[[]T]) numberSliceField[T] {
	return numberSliceField[T]{field: f.field.withRule(rule)}
}

func (f numberSliceField[T]) Required() numberSliceField[T]      { return f.withRule(requiredSlice[T]) }
func (f numberSliceField[T]) MinItems(n int) numberSliceField[T] { return f.withRule(minItems[T](n)) }
func (f numberSliceField[T]) MaxItems(n int) numberSliceField[T] { return f.withRule(maxItems[T](n)) }

// Each transitions to a per-element builder exposing number rules.
func (f numberSliceField[T]) Each() numberSliceEach[T] {
	return numberSliceEach[T]{eachBuilder[T]{name: f.name, value: f.value, lengthRules: f.rules}}
}

type numberSliceEach[T number] struct {
	eachBuilder[T]
}

func (e numberSliceEach[T]) with(rule Rule[T]) numberSliceEach[T] {
	return numberSliceEach[T]{e.withElem(rule)}
}

func (e numberSliceEach[T]) Required() numberSliceEach[T]           { return e.with(requiredNumber[T]) }
func (e numberSliceEach[T]) Min(n T) numberSliceEach[T]             { return e.with(minValue(n)) }
func (e numberSliceEach[T]) Max(n T) numberSliceEach[T]             { return e.with(maxValue(n)) }
func (e numberSliceEach[T]) Custom(rule Rule[T]) numberSliceEach[T] { return e.with(rule) }

// --- slice length rules ---

func requiredSlice[T any](v []T) error {
	if len(v) == 0 {
		return errors.New("is required")
	}
	return nil
}

func minItems[T any](n int) Rule[[]T] {
	return func(v []T) error {
		if len(v) < n {
			return fmt.Errorf("must contain at least %d items", n)
		}
		return nil
	}
}

func maxItems[T any](n int) Rule[[]T] {
	return func(v []T) error {
		if len(v) > n {
			return fmt.Errorf("must contain no more than %d items", n)
		}
		return nil
	}
}
