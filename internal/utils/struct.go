package utils

import (
	"fmt"
	"reflect"
	"strings"
)

func StructKeys(s any, tagName string) []string {
	keys := []string{}
	ct := reflect.TypeOf(s)

	if ct.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < ct.NumField(); i++ {
		field := ct.Field(i)
		tag := field.Tag.Get(tagName)
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		if field.Type.Kind() == reflect.Struct {
			res := StructKeys(reflect.New(field.Type).Elem().Interface(), tagName)
			for _, k := range res {
				keys = append(keys, fmt.Sprintf("%s.%s", tag, k))
			}
		} else {
			keys = append(keys, tag)
		}
	}

	return keys
}

// StructToMap serializes a struct (or pointer to struct) into a map,
// including only fields marked with the `audit` tag.
func StructToMap(input any) map[string]any {
	v := reflect.ValueOf(input)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return map[string]any{}
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return map[string]any{}
	}

	t := v.Type()
	out := map[string]any{}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}

		tag, ok := sf.Tag.Lookup("audit")
		if !ok || tag == "" || tag == "-" {
			continue
		}
		key, options, _ := strings.Cut(tag, ",")
		if key == "" || key == "-" {
			continue
		}

		fv := v.Field(i)
		if hasTagOption(options, "omitempty") && isEmptyValue(fv) {
			continue
		}

		if fv.Kind() == reflect.Pointer {
			if fv.IsNil() {
				continue
			}
			out[key] = fv.Elem().Interface()
			continue
		}

		out[key] = fv.Interface()
	}

	return out
}

func hasTagOption(options string, want string) bool {
	if options == "" {
		return false
	}
	for _, opt := range strings.Split(options, ",") {
		if strings.TrimSpace(opt) == want {
			return true
		}
	}
	return false
}

// isEmptyValue mirrors encoding/json emptiness checks for omitempty.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	case reflect.Struct:
		return v.IsZero()
	}

	return false
}
