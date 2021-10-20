package rs

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

func (TypeHintObjects) apply(v *alterInterface) (*alterInterface, error) {
	var rawBytes []byte
	switch vt := v.scalarData.(type) {
	case []byte:
		rawBytes = vt
	case string:
		rawBytes = []byte(vt)
	default:
		rv := reflect.ValueOf(v.Value())

	convertToObjects:
		switch vk := rv.Kind(); vk {
		case reflect.Invalid:
			return nil, nil
		case reflect.Array, reflect.Slice, reflect.Interface:
			return v, nil
		case reflect.Ptr:
			rv = rv.Elem()
			goto convertToObjects
		default:
			return nil, fmt.Errorf(
				"incompatible type %T as object array",
				vt,
			)
		}
	}

	ret := new(alterInterface)
	err := yaml.Unmarshal(rawBytes, &ret.sliceData)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal bytes %q as object array: %w",
			string(rawBytes), err,
		)
	}
	return ret, nil
}
