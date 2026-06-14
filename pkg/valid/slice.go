package valid

import (
	"errors"
	"fmt"
)

// Slice validates a slice field. Rules operate on the whole slice (e.g.
// NotEmpty, MinItems); use Each to apply per-element rules.
func Slice[T any](name string, value []T, rules ...Rule[[]T]) Field {
	return newField(name, value, rules)
}

// NotEmpty asserts the slice has at least one element.
func NotEmpty[T any](v []T) error {
	if len(v) == 0 {
		return errors.New("is required")
	}
	return nil
}

// Each wraps per-element rules into a single slice rule, reporting
// "element #i: <problem>" on the first failing element.
func Each[T any](rules ...Rule[T]) Rule[[]T] {
	return func(vs []T) error {
		for i, v := range vs {
			for _, rule := range rules {
				if err := rule(v); err != nil {
					return fmt.Errorf("element #%d: %w", i, err)
				}
			}
		}
		return nil
	}
}
