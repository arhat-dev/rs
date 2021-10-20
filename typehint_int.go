package rs

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

func (thi TypeHintInt) apply(in *alterInterface) (*alterInterface, error) {
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
				thi, rv.Type().String(), err,
			)
		}

		rawStr = string(dataBytes)
	} else {
		switch vt := rv.Interface().(type) {
		case float32:
			in.scalarData = int64(vt)
		case float64:
			in.scalarData = int64(vt)
		case int:
			in.scalarData = vt
		case int8:
			in.scalarData = vt
		case int16:
			in.scalarData = vt
		case int32:
			in.scalarData = vt
		case int64:
			in.scalarData = vt
		case uint:
			in.scalarData = int64(vt)
		case uint8:
			in.scalarData = int64(vt)
		case uint16:
			in.scalarData = int64(vt)
		case uint32:
			in.scalarData = int64(vt)
		case uint64:
			in.scalarData = int64(vt)
		case uintptr:
			in.scalarData = int64(vt)
		case []byte:
			rawStr = string(vt)
			in.scalarData = nil
		case string:
			rawStr = vt
			in.scalarData = nil
		default:
			return nil, incompatibleTypeError(thi, in.scalarData)
		}
	}

	if in.scalarData != nil {
		return in, nil
	}

	strV, err := strconv.Unquote(rawStr)
	if err != nil {
		// may fail when not quoted
		strV = rawStr
	}

	intV, err := strconv.ParseInt(strV, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"typehint.%s: failed to unmarshal %q as int: %w",
			thi, rawStr, err,
		)
	}

	return &alterInterface{
		scalarData:   intV,
		originalNode: in.originalNode,
	}, nil
}
