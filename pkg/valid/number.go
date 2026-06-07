package valid

import (
	"errors"
	"fmt"
)

// number constrains Number to integer and floating-point kinds. Rules use the
// generic operators directly, so values are never widened (no int64 coercion)
// and large unsigned values stay correct.
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// numberField validates a numeric value.
type numberField[T number] struct {
	field[T]
}

// Number validates a numeric field.
func Number[T number](name string, value T) numberField[T] {
	return numberField[T]{field: newField(name, value)}
}

func (f numberField[T]) withRule(rule Rule[T]) numberField[T] {
	return numberField[T]{field: f.field.withRule(rule)}
}

func (f numberField[T]) Required() numberField[T]           { return f.withRule(requiredNumber[T]) }
func (f numberField[T]) Min(n T) numberField[T]             { return f.withRule(minValue(n)) }
func (f numberField[T]) Max(n T) numberField[T]             { return f.withRule(maxValue(n)) }
func (f numberField[T]) Custom(rule Rule[T]) numberField[T] { return f.withRule(rule) }

// --- number rules ---

func requiredNumber[T number](v T) error {
	if v == 0 {
		return errors.New("is required")
	}
	return nil
}

func minValue[T number](n T) Rule[T] {
	return func(v T) error {
		if v < n {
			return fmt.Errorf("must be >= %v", n)
		}
		return nil
	}
}

func maxValue[T number](n T) Rule[T] {
	return func(v T) error {
		if v > n {
			return fmt.Errorf("must be <= %v", n)
		}
		return nil
	}
}
