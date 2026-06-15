package valid

// Struct validates a nested Validatable, namespacing its problems under name
// (name.<child key>), so nesting composes to arbitrary depth (a.b.c). Pass a
// value whose Validate() satisfies Validatable — e.g. &in.Inner when Validate
// has a pointer receiver. A nil value is skipped, so an optional nested struct
// can be passed directly. Note: a typed nil pointer is not a nil interface — a
// caller with an optional *Inner field should pass a genuine nil.
func Struct(name string, value Validatable) Validatable {
	return structField{name: name, value: value}
}

// structField adapts a nested Validatable, prefixing its keys with the field
// name.
type structField struct {
	name  string
	value Validatable
}

func (f structField) Validate() Problems {
	if f.value == nil {
		return nil
	}
	return prefixed(f.name, f.value.Validate())
}
