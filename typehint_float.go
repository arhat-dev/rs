package rs

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

func (thf TypeHintFloat) apply(in *alterInterface) (*alterInterface, error) {
	rv := reflect.ValueOf(in.scalarData)

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return nil, nil
	}

	var rawStr string

	if rv.Kind() == reflect.Interface {
		dataBytes, err := yaml.Marshal(rv.Interface())
		if err != nil {
			return nil, fmt.Errorf(
				"typehint.%s: failed marshaling interface value of %q to yaml bytes: %w",
				thf, rv.Type().String(), err,
			)
		}

		rawStr = string(dataBytes)
	} else {
		switch vt := rv.Interface().(type) {
		case float32:
			in.scalarData = vt
		case float64:
			in.scalarData = vt
		case int:
			in.scalarData = float64(vt)
		case int8:
			in.scalarData = float64(vt)
		case int16:
			in.scalarData = float64(vt)
		case int32:
			in.scalarData = float64(vt)
		case int64:
			in.scalarData = float64(vt)
		case uint:
			in.scalarData = float64(vt)
		case uint8:
			in.scalarData = float64(vt)
		case uint16:
			in.scalarData = float64(vt)
		case uint32:
			in.scalarData = float64(vt)
		case uint64:
			in.scalarData = float64(vt)
		case uintptr:
			in.scalarData = float64(vt)
		case []byte:
			rawStr = string(vt)
			in.scalarData = nil
		case string:
			rawStr = vt
			in.scalarData = nil
		default:
			return nil, incompatibleTypeError(thf, in.scalarData)
		}
	}

	if in.scalarData != nil {
		return in, nil
	}

	strV, err := strconv.Unquote(rawStr)
	if err != nil {
		strV = rawStr
	}

	f64v, err := strconv.ParseFloat(strV, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"typehint.%s: failed to unmarshal string %q as float: %w",
			thf, rawStr, err,
		)
	}

	return &alterInterface{
		scalarData:   f64v,
		originalNode: in.originalNode,
	}, nil
}
