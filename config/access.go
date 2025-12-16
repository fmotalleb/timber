package config

import (
	"fmt"
	"reflect"
	"strings"
)

type Access struct {
	Paths []string `mapstructure:"path"`
}

func (a *Access) Decode(from reflect.Type, val interface{}) (any, error) {
	switch from.Kind() {

	case reflect.String:
		raw := val.(string)
		a.Paths = strings.Split(raw, ",")
		return *a, nil

	case reflect.Slice:
		rawSlice, ok := val.([]interface{})
		if !ok {
			return val, nil
		}

		paths := make([]string, 0, len(rawSlice))
		for i, v := range rawSlice {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("access[%d] is not a string", i)
			}
			paths = append(paths, s)
		}

		a.Paths = paths
		return *a, nil

	default:
		return val, nil
	}
}
