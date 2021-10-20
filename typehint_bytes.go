package rs

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// convert to bytes, only return error when marshaling failed
func (thb TypeHintBytes) apply(in *alterInterface) (*alterInterface, error) {
	if in.originalNode != nil && in.originalNode.Kind == yaml.ScalarNode {
		return &alterInterface{
			scalarData:   []byte(in.originalNode.Value),
			originalNode: in.originalNode,
		}, nil
	}

	rv := reflect.ValueOf(in.scalarData)

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return nil, nil
	}

	switch vt := rv.Interface().(type) {
	case string:
		return &alterInterface{
			scalarData:   []byte(vt),
			originalNode: in.originalNode,
		}, nil
	case []byte:
		return &alterInterface{
			scalarData:   vt,
			originalNode: in.originalNode,
		}, nil
	default:
		// println(
		// 	"RV.typ", fmt.Sprintf("%T", rv.Interface()),
		// 	"IN.typ", fmt.Sprintf("%T", in.Value()),
		// )
		vt = in.Value()
		dataBytes, err := yaml.Marshal(vt)
		if err != nil {
			return nil, fmt.Errorf(
				"typehint.%s: failed to marshal %T to yaml bytes: %w",
				thb, vt, err,
			)
		}

		return &alterInterface{
			scalarData:   dataBytes,
			originalNode: in.originalNode,
		}, nil
	}
}
