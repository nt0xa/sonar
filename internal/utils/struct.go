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

// TODO: remove when event meta is coverted to struct
func StructToMap(input any) map[string]any {
	out := make(map[string]any)
	val := reflect.ValueOf(input)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return out
		}
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// Skip unexported fields
		if typeField.PkgPath != "" {
			continue
		}

		name := typeField.Name
		tag := typeField.Tag.Get("json")
		if tag != "" && tag != "-" {
			name = tag
		}

		if isZero(field) {
			continue
		}

		switch field.Kind() {
		case reflect.Struct:
			m := StructToMap(field.Interface())
			if len(m) > 0 {
				out[name] = m
			}
		case reflect.Ptr:
			if !field.IsNil() {
				if field.Elem().Kind() == reflect.Struct {
					m := StructToMap(field.Elem().Interface())
					if len(m) > 0 {
						out[name] = m
					}
				} else if !isZero(field.Elem()) {
					out[name] = field.Elem().Interface()
				}
			}
		case reflect.Slice, reflect.Array:
			if field.Len() > 0 {
				slice := make([]any, 0, field.Len())
				for j := 0; j < field.Len(); j++ {
					v := field.Index(j)
					if v.Kind() == reflect.Struct {
						m := StructToMap(v.Interface())
						if len(m) > 0 {
							slice = append(slice, m)
						}
					} else {
						if !isZero(v) {
							slice = append(slice, v.Interface())
						}
					}
				}
				if len(slice) > 0 {
					out[name] = slice
				}
			}
		case reflect.Map:
			if field.Len() > 0 {
				mapVal := make(map[string]any)
				for _, key := range field.MapKeys() {
					v := field.MapIndex(key)
					if v.Kind() == reflect.Struct {
						m := StructToMap(v.Interface())
						if len(m) > 0 {
							mapVal[fmt.Sprint(key.Interface())] = m
						}
					} else {
						if !isZero(v) {
							mapVal[fmt.Sprint(key.Interface())] = v.Interface()
						}
					}
				}
				if len(mapVal) > 0 {
					out[name] = mapVal
				}
			}
		default:
			out[name] = field.Interface()
		}
	}

	return out
}

// isZero checks if a reflect.Value is zero value
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice, reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	}
	zero := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), zero.Interface())
}
