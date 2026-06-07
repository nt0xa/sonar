// Package valid provides declarative, type-safe struct validation.
//
// Validation is expressed as a method on the params type that returns a map of
// field name -> problem (or nil when valid):
//
//	func (in CreateInput) Validate() map[string]string {
//		return v.Struct(&in,
//			v.String(&in.Name, v.Required, v.MinLength(3)),
//			v.Int(&in.TTL, v.Required, v.Max(100)),
//		)
//	}
//
// Type safety comes from one interface per value category: string-only rules
// (MinLength, Match, ...) cannot be passed to Int, number-only rules (Max)
// cannot be passed to String, and Each cannot be passed to a non-string Slice.
// Cross-type rules such as Required work on any field.
//
// Struct, the field constructors, and the generic rules are domain-agnostic.
// Domain-specific validators (e.g. a subdomain check) are plain
// func(string) error values adapted with By.
package valid

import (
	"reflect"
	"strings"
	"unicode"
)

// signed constrains Int to signed integer kinds (incl. enums). Unsigned types
// are intentionally excluded: widening a large uint64 to int64 would corrupt the
// value, so an Uint constructor should be added when unsigned fields need rules.
type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Problems maps a json-style field name to its validation problem. It is a
// defined type (not an alias) so it reads as a domain type at call sites and
// leaves room to grow (e.g. map[string]Problem) without churning signatures; it
// still marshals to JSON exactly like its underlying map, and is assignable to a
// plain map[string]string parameter. The zero/nil value means "no problems".
type Problems map[string]string

// Field binds a struct field pointer to its rules. It is produced by String,
// Int, Slice, etc. and consumed by Struct.
type Field struct {
	ptr   any
	check func() string
}

// Struct validates the fields of the struct pointed to by s and returns a map of
// json-style field name -> problem. It returns nil when everything is valid.
//
// Struct panics on programmer misuse: s must be a non-nil pointer to a struct,
// and every field pointer must point into that struct.
func Struct(s any, fields ...Field) Problems {
	rv := reflect.ValueOf(s)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		panic("valid: Struct requires a non-nil pointer to a struct")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		panic("valid: Struct requires a non-nil pointer to a struct")
	}
	rt := rv.Type()

	problems := make(Problems)
	for _, f := range fields {
		if msg := f.check(); msg != "" {
			problems[fieldName(rv, rt, f.ptr)] = msg
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return problems
}

// String validates a string (or ~string, e.g. an enum) field.
func String[T ~string](ptr *T, rules ...StringRule) Field {
	return Field{ptr: ptr, check: func() string {
		return applyString(string(*ptr), rules)
	}}
}

// OptionalStringPtr validates an optional string (or ~string) field. When the
// pointer is nil the field is treated as absent and all rules are skipped.
func OptionalStringPtr[T ~string](ptr **T, rules ...StringRule) Field {
	return Field{ptr: ptr, check: func() string {
		if *ptr == nil {
			return ""
		}
		return applyString(string(**ptr), rules)
	}}
}

// Int validates a signed integer field.
func Int[T signed](ptr *T, rules ...NumberRule) Field {
	return Field{ptr: ptr, check: func() string {
		n := int64(*ptr)
		for _, r := range rules {
			if msg := r.checkNumber(n); msg != "" {
				return msg
			}
		}
		return ""
	}}
}

// Slice validates a slice of any element type with length rules (Required,
// MinItems, MaxItems). For per-element rules on string slices use StringSlice.
func Slice[T any](ptr *[]T, rules ...SliceRule) Field {
	return Field{ptr: ptr, check: func() string {
		n := len(*ptr)
		for _, r := range rules {
			if msg := r.checkSlice(n); msg != "" {
				return msg
			}
		}
		return ""
	}}
}

// StringSlice validates a slice of ~string values. It accepts length rules and
// Each (per-element string rules).
func StringSlice[T ~string](ptr *[]T, rules ...StringSliceRule) Field {
	return Field{ptr: ptr, check: func() string {
		elems := make([]string, len(*ptr))
		for i, e := range *ptr {
			elems[i] = string(e)
		}
		for _, r := range rules {
			if msg := r.checkStringSlice(elems); msg != "" {
				return msg
			}
		}
		return ""
	}}
}

// applyString runs string rules in order. When Optional is present and the value
// is empty, the remaining rules are skipped.
func applyString(s string, rules []StringRule) string {
	for _, r := range rules {
		if _, ok := r.(optionalRule); ok && s == "" {
			return ""
		}
	}
	for _, r := range rules {
		if _, ok := r.(optionalRule); ok {
			continue
		}
		if msg := r.checkString(s); msg != "" {
			return msg
		}
	}
	return ""
}

// fieldName finds the struct field whose address matches ptr and returns its
// external name: the json tag if present, otherwise lowerCamel of the Go name.
// It panics if ptr does not point into the struct (programmer misuse).
func fieldName(rv reflect.Value, rt reflect.Type, ptr any) string {
	target := reflect.ValueOf(ptr).Pointer()
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).Addr().Pointer() == target {
			if tag := rt.Field(i).Tag.Get("json"); tag != "" && tag != "-" {
				return strings.Split(tag, ",")[0]
			}
			return lowerCamel(rt.Field(i).Name)
		}
	}
	panic("valid: field pointer does not point into the validated struct")
}

// lowerCamel converts a Go field name to its lowerCamel json equivalent,
// handling leading acronyms: Name -> name, PayloadName -> payloadName,
// TTL -> ttl, ID -> id, APIToken -> apiToken.
func lowerCamel(s string) string {
	r := []rune(s)
	n := 0
	for n < len(r) && unicode.IsUpper(r[n]) {
		n++
	}
	switch {
	case n == 0:
		return s
	case n == len(r):
		// All uppercase: TTL -> ttl, ID -> id.
		return strings.ToLower(s)
	case n == 1:
		// Single leading uppercase: PayloadName -> payloadName.
		return string(unicode.ToLower(r[0])) + string(r[1:])
	default:
		// Acronym followed by a word: APIToken -> apiToken. The last uppercase
		// rune starts the next word.
		return strings.ToLower(string(r[:n-1])) + string(r[n-1:])
	}
}
