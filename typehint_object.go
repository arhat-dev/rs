package rs

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

func (tho TypeHintObject) apply(v *alterInterface) (*alterInterface, error) {
	rv := reflect.ValueOf(v.Value())

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return nil, nil
	}

	var rawBytes []byte
	switch vk := rv.Kind(); vk {
	case reflect.Map, reflect.Struct:
		return v, nil
	case reflect.Interface:
		rv = rv.Elem()
		if !rv.IsValid() {
			return nil, nil
		}

		// switch vk = rv.Kind(); {
		// case vk == reflect.Map, vk == reflect.Struct:
		// 	return v, nil
		// default:
		// 	switch vt := rv.Interface().(type) {
		// 	case string:
		// 	case []byte:
		// 	}
		// }
	default:
		switch vt := rv.Interface().(type) {
		case []byte:
			rawBytes = vt
		case string:
			rawBytes = []byte(vt)
		default:
			return nil, incompatibleTypeError(tho, v.scalarData)
		}
	}

	ret := new(alterInterface)
	err := yaml.Unmarshal(rawBytes, &ret.mapData)
	if err != nil {
		return nil, fmt.Errorf(
			"typehint.%s: failed to unmarshal bytes %q as object: %w",
			tho, string(rawBytes), err,
		)
	}

	return ret, nil
}
