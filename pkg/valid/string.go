package valid

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// String validates a string (or ~string, e.g. an enum) field.
func String[T ~string](name string, value T, rules ...Rule[T]) Validatable {
	return newField(name, value, rules)
}

// OptionalString validates an optional string field addressed by a pointer. When
// the pointer is nil the field is treated as absent (no rules run).
func OptionalString[T ~string](name string, value *T, rules ...Rule[T]) Validatable {
	if value == nil {
		var zero T
		return newField(name, zero, nil)
	}
	return newField(name, *value, rules)
}

// NotBlank asserts the string is not empty after trimming whitespace.
func NotBlank[T ~string](v T) error {
	if strings.TrimSpace(string(v)) == "" {
		return errors.New("must not be blank")
	}
	return nil
}

// MinLength asserts the string has at least n characters.
func MinLength(n int) Rule[string] {
	return func(v string) error {
		if utf8.RuneCountInString(v) < n {
			return fmt.Errorf("must be at least %d characters", n)
		}
		return nil
	}
}

// MaxLength asserts the string has at most n characters.
func MaxLength(n int) Rule[string] {
	return func(v string) error {
		if utf8.RuneCountInString(v) > n {
			return fmt.Errorf("must be at most %d characters", n)
		}
		return nil
	}
}

// Match asserts the string matches re, reporting msg on failure.
func Match(re *regexp.Regexp, msg string) Rule[string] {
	return func(v string) error {
		if !re.MatchString(v) {
			return errors.New(msg)
		}
		return nil
	}
}

// In asserts the value equals one of the allowed values. It accepts ~string
// values (e.g. enum value slices via In(TypeValues()...)) and is usable inside
// Each.
func In[T ~string](allowed ...T) Rule[T] {
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
