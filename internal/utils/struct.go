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
