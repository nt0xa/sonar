package valid

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// stringField validates a required (always-present) string or ~string value.
type stringField[T ~string] struct {
	field[T]
}

// String validates a string (or ~string, e.g. an enum) field.
func String[T ~string](name string, value T) stringField[T] {
	return stringField[T]{field: newField(name, value)}
}

func (f stringField[T]) withRule(rule Rule[T]) stringField[T] {
	return stringField[T]{field: f.field.withRule(rule)}
}

func (f stringField[T]) Required() stringField[T]       { return f.withRule(requiredString[T]) }
func (f stringField[T]) NotBlank() stringField[T]       { return f.withRule(notBlank[T]) }
func (f stringField[T]) MinLength(n int) stringField[T] { return f.withRule(minLength[T](n)) }
func (f stringField[T]) MaxLength(n int) stringField[T] { return f.withRule(maxLength[T](n)) }
func (f stringField[T]) In(allowed ...T) stringField[T] { return f.withRule(OneOf(allowed...)) }

func (f stringField[T]) Match(re *regexp.Regexp, msg string) stringField[T] {
	return f.withRule(match[T](re, msg))
}

func (f stringField[T]) Custom(rule Rule[T]) stringField[T] { return f.withRule(rule) }

// optionalStringField validates an optional string via a pointer: when the
// pointer is nil the field is treated as absent and all rules are skipped. This
// is the only supported form of string optionality — a value-type string is
// always present and must satisfy its rules.
type optionalStringField[T ~string] struct {
	name  string
	value *T
	rules []Rule[T]
}

// OptionalString validates an optional string (or ~string) field addressed by a
// pointer.
func OptionalString[T ~string](name string, value *T) optionalStringField[T] {
	return optionalStringField[T]{name: name, value: value}
}

func (f optionalStringField[T]) withRule(rule Rule[T]) optionalStringField[T] {
	rules := make([]Rule[T], 0, len(f.rules)+1)
	rules = append(rules, f.rules...)
	rules = append(rules, rule)

	return optionalStringField[T]{name: f.name, value: f.value, rules: rules}
}

func (f optionalStringField[T]) Validate() (string, error) {
	if f.value == nil {
		return f.name, nil
	}

	for _, rule := range f.rules {
		if err := rule(*f.value); err != nil {
			return f.name, err
		}
	}

	return f.name, nil
}

func (f optionalStringField[T]) NotBlank() optionalStringField[T] { return f.withRule(notBlank[T]) }
func (f optionalStringField[T]) MinLength(n int) optionalStringField[T] {
	return f.withRule(minLength[T](n))
}
func (f optionalStringField[T]) MaxLength(n int) optionalStringField[T] {
	return f.withRule(maxLength[T](n))
}
func (f optionalStringField[T]) In(allowed ...T) optionalStringField[T] {
	return f.withRule(OneOf(allowed...))
}

func (f optionalStringField[T]) Match(re *regexp.Regexp, msg string) optionalStringField[T] {
	return f.withRule(match[T](re, msg))
}

func (f optionalStringField[T]) Custom(rule Rule[T]) optionalStringField[T] { return f.withRule(rule) }

// --- string rules ---

func requiredString[T ~string](v T) error {
	if v == "" {
		return errors.New("is required")
	}
	return nil
}

func notBlank[T ~string](v T) error {
	if strings.TrimSpace(string(v)) == "" {
		return errors.New("must not be blank")
	}
	return nil
}

func minLength[T ~string](n int) Rule[T] {
	return func(v T) error {
		if utf8.RuneCountInString(string(v)) < n {
			return fmt.Errorf("must be at least %d characters", n)
		}
		return nil
	}
}

func maxLength[T ~string](n int) Rule[T] {
	return func(v T) error {
		if utf8.RuneCountInString(string(v)) > n {
			return fmt.Errorf("must be at most %d characters", n)
		}
		return nil
	}
}

func match[T ~string](re *regexp.Regexp, msg string) Rule[T] {
	return func(v T) error {
		if !re.MatchString(string(v)) {
			return errors.New(msg)
		}
		return nil
	}
}

// OneOf returns a rule asserting the value equals one of the allowed values. It
// is also usable standalone (e.g. inside Each).
func OneOf[T ~string](allowed ...T) Rule[T] {
	return func(v T) error {
		if slices.Contains(allowed, v) {
			return nil
		}
		strs := make([]string, len(allowed))
		for i, a := range allowed {
			strs[i] = string(a)
		}
		return fmt.Errorf("must be one of: %s", strings.Join(strs, ", "))
	}
}
