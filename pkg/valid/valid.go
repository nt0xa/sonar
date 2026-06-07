// Package valid provides declarative, type-safe struct validation with a fluent
// builder API:
//
//	func (in CreateInput) Validate() valid.Problems {
//		return valid.Struct(
//			valid.String("name", in.Name).Required().MinLength(3),
//			valid.Number("ttl", in.TTL).Max(100),
//			valid.StringSlice("values", in.Values).Required().Each(isIPv4),
//		)
//	}
//
// Each field carries its own json-style name (passed explicitly, no reflection)
// and an ordered list of rules. Type safety comes from per-type field builders:
// string rules live on String/OptionalString, numeric rules on Number, and
// per-element rules on StringSlice — so a numeric rule cannot be attached to a
// string field, and Each cannot be attached to a non-string slice.
package valid

// Field is a single validated struct field. Struct collects the problem reported
// by each field.
type Field interface {
	Validate() (name string, err error)
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

// Struct validates the given fields and returns a map of field name -> problem.
// Within a field the first failing rule wins; if two fields share a name the
// first is kept. It returns nil when everything is valid.
func Struct(fields ...Field) Problems {
	problems := Problems{}

	for _, field := range fields {
		name, err := field.Validate()
		if err != nil {
			if _, exists := problems[name]; !exists {
				problems[name] = err.Error()
			}
		}
	}

	if len(problems) == 0 {
		return nil
	}

	return problems
}

// field is the shared implementation for all typed field builders: a name, a
// value, and an ordered list of rules.
type field[T any] struct {
	name  string
	value T
	rules []Rule[T]
}

func newField[T any](name string, value T) field[T] {
	return field[T]{name: name, value: value}
}

func (f field[T]) withRule(rule Rule[T]) field[T] {
	rules := make([]Rule[T], 0, len(f.rules)+1)
	rules = append(rules, f.rules...)
	rules = append(rules, rule)

	return field[T]{name: f.name, value: f.value, rules: rules}
}

func (f field[T]) Validate() (string, error) {
	for _, rule := range f.rules {
		if err := rule(f.value); err != nil {
			return f.name, err
		}
	}

	return f.name, nil
}
