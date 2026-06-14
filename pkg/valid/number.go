package valid

import "fmt"

// number constrains Number to integer and floating-point kinds. Rules use the
// generic operators directly, so values are never widened and large unsigned
// values stay correct.
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Number validates a numeric field.
func Number[T number](name string, value T, rules ...Rule[T]) Field {
	return newField(name, value, rules)
}

// Min asserts the number is at least n.
func Min[T number](n T) Rule[T] {
	return func(v T) error {
		if v < n {
			return fmt.Errorf("must be >= %v", n)
		}
		return nil
	}
}

// Max asserts the number is at most n.
func Max[T number](n T) Rule[T] {
	return func(v T) error {
		if v > n {
			return fmt.Errorf("must be <= %v", n)
		}
		return nil
	}
}
