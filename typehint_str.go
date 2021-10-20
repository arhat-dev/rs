package rs

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// convert to bytes, only return error when marshaling failed
func (ths TypeHintStr) apply(in *alterInterface) (*alterInterface, error) {
	if in.originalNode != nil && in.originalNode.Kind == yaml.ScalarNode {
		return &alterInterface{
			scalarData:   in.originalNode.Value,
			originalNode: in.originalNode,
		}, nil
	}

	rv := reflect.ValueOf(in.Value())

	// derefernce to check []byte, and string
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return nil, nil
	}

	switch vt := rv.Interface().(type) {
	case []byte:
		return &alterInterface{
			scalarData:   string(vt),
			originalNode: in.originalNode,
		}, nil
	case string:
		return &alterInterface{
			scalarData:   vt,
			originalNode: in.originalNode,
		}, nil
	default:
		vt = in.Value()
		bytesV, err := yaml.Marshal(vt)
		if err != nil {
			return nil, fmt.Errorf(
				"typehint.%s: failed to marshal %T value as str",
				ths, vt,
			)
		}

		return &alterInterface{
			scalarData:   string(bytesV),
			originalNode: in.originalNode,
		}, nil
	}
}
