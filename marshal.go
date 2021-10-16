//go:build !rs_noyamlmarshaler
// +build !rs_noyamlmarshaler

package rs

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

var _ yaml.Marshaler = (*BaseField)(nil)

func (f *BaseField) MarshalYAML() (interface{}, error) {
	if atomic.LoadUint32(&f._initialized) != 1 {
		return nil, fmt.Errorf("rs: struct not intialized before marshaling")
	}

	var ret map[string]interface{}
	for k, v := range f.fields {
		if v.omitempty {
			switch v.fieldValue.Kind() {
			case reflect.Invalid:
				continue
			case reflect.Slice, reflect.Array, reflect.Map:
				if v.fieldValue.Len() == 0 {
					continue
				}
			case reflect.Ptr:
				if v.fieldValue.IsNil() {
					continue
				}
			default:
				if v.fieldValue.IsZero() {
					continue
				}
			}
		}

		if ret == nil {
			ret = make(map[string]interface{})
		}

		ret[k] = v.fieldValue.Interface()
	}

	if ret == nil && !f._parentValue.IsValid() {
		return nil, nil
	}

	return ret, nil
}