package rs

import (
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

func (f *BaseField) HasUnresolvedField() bool {
	return len(f.unresolvedFields) != 0
}

func (f *BaseField) ResolveFields(rc RenderingHandler, depth int, fieldNames ...string) error {
	if atomic.LoadUint32(&f._initialized) == 0 {
		return fmt.Errorf("rs: struct not intialized before resolving")
	}

	if depth == 0 {
		return nil
	}

	parentStructType := f._parentValue.Type()

	if len(fieldNames) == 0 {
		// resolve all

		for i := 1; i < f._parentValue.NumField(); i++ {
			sf := parentStructType.Field(i)
			if len(sf.PkgPath) != 0 {
				// not exported
				continue
			}

			err := f.resolveSingleField(
				rc, depth, sf.Name, f._parentValue.Field(i),
			)
			if err != nil {
				return fmt.Errorf(
					"rs: failed to resolve field %s.%s: %w",
					parentStructType.String(), sf.Name, err,
				)
			}
		}

		return nil
	}

	for _, name := range fieldNames {
		fv := f._parentValue.FieldByName(name)
		if !fv.IsValid() {
			return fmt.Errorf("rs: no such field %q in struct %q", name, parentStructType.String())
		}

		err := f.resolveSingleField(rc, depth, name, fv)
		if err != nil {
			return fmt.Errorf(
				"rs: failed to resolve requested field %s.%s: %w",
				parentStructType.String(), name, err,
			)
		}
	}

	return nil
}

func (f *BaseField) resolveSingleField(
	rc RenderingHandler,
	depth int,
	fieldName string,
	targetField reflect.Value,
) error {
	keepOld := false

	for k, v := range f.unresolvedFields {
		if v.fieldName == fieldName {
			err := f.handleUnResolvedField(
				rc, depth, k, v, keepOld,
			)
			if err != nil {
				return err
			}

			keepOld = true
		}
	}

	return f.handleResolvedField(rc, depth, targetField)
}

// nolint:gocyclo
func (f *BaseField) handleResolvedField(
	rc RenderingHandler,
	depth int,
	field reflect.Value,
) error {
	if depth == 0 {
		return nil
	}

	switch field.Kind() {
	case reflect.Map:
		if field.IsNil() {
			return nil
		}

		iter := field.MapRange()
		for iter.Next() {
			err := f.handleResolvedField(rc, depth-1, iter.Value())
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if field.IsNil() {
			// this is a resolved field, slice/array empty means no value
			return nil
		}

		for i := 0; i < field.Len(); i++ {
			tt := field.Index(i)
			err := f.handleResolvedField(rc, depth-1, tt)
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		// handled after switch
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		if !field.IsValid() || field.IsZero() || field.IsNil() {
			return nil
		}
		// handled after switch
	default:
		// scalar types, no action required
		return nil
	}

	return tryResolve(rc, depth, field)
}

func tryResolve(rc RenderingHandler, depth int, targetField reflect.Value) error {
	if targetField.CanInterface() {
		fVal, canCallResolve := targetField.Interface().(Field)
		if canCallResolve {
			return fVal.ResolveFields(rc, depth)
		}
	}

	if targetField.CanAddr() && targetField.Addr().CanInterface() {
		fVal, canCallResolve := targetField.Addr().Interface().(Field)
		if canCallResolve {
			return fVal.ResolveFields(rc, depth)
		}
	}

	// no more field to resolve
	return nil
}

func (f *BaseField) handleUnResolvedField(
	rc RenderingHandler,
	depth int,
	key unresolvedFieldKey,
	v *unresolvedFieldValue,
	keepOld bool,
) error {
	target := v.fieldValue

	for i, rawData := range v.rawDataList {
		toResolve := rawData
		if v.isCatchOtherField {
			// unwrap map data for resolving
			toResolve = rawData.mapData[key.yamlKey]
		}

		var (
			resolvedValue interface{}

			err error
		)

		for _, renderer := range v.renderers {
			// a patch is implied when the renderer has a `!` suffix
			var patchSpec *renderingPatchSpec
			if renderer.patchSpec {
				var valueToPatch interface{}
				patchSpec, valueToPatch, err = f.resolvePatchSpec(rc, toResolve.NormalizedValue())
				if err != nil {
					return fmt.Errorf(
						"failed to resolve patch spec for renderer %q: %w",
						renderer.name, err,
					)
				}

				toResolve = &alterInterface{
					scalarData: valueToPatch,
				}
			}

			// apply hint before resolving (rendering)
			hint := renderer.typeHint

			// TBD: shall we add hint for typed data?
			// 		 seems not necessary since use yaml.Unmarshal can handle it
			//
			// 			if hint == TypeHintNone {
			// 				ref := target.Type()
			// 				for ref.Kind() == reflect.Ptr {
			// 					ref = ref.Elem()
			// 				}
			//
			// 				switch ref.Kind() {
			// 				case reflect.String:
			// 					hint = TypeHintStr
			// 				case reflect.Slice:
			// 					if ref.Elem().Kind() == reflect.Uint8 {
			// 						hint = TypeHintBytes
			// 					}
			// 				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			// 					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			// 					reflect.Uintptr:
			// 					hint = TypeHintInt
			// 				case reflect.Float32, reflect.Float64:
			// 					hint = TypeHintFloat
			// 				case reflect.Struct, reflect.Map:
			// 					hint = TypeHintMap
			// 				default:
			// 					// no hint
			// 				}
			// 			}

			resolvedValue, err = applyTypeHint(hint, toResolve)
			if err != nil {
				return fmt.Errorf(
					"failed to ensure type hint %q on yaml key %q: %w",
					hint, key.yamlKey, err,
				)
			}

			if len(renderer.name) != 0 {
				resolvedValue, err = rc.RenderYaml(renderer.name, resolvedValue)
				if err != nil {
					return fmt.Errorf(
						"renderer %q failed to render value: %w",
						renderer.name, err,
					)
				}
			}

			if patchSpec != nil {
				resolvedValue, err = patchSpec.ApplyTo(resolvedValue)
				if err != nil {
					return fmt.Errorf(
						"failed to apply patches: %w",
						err,
					)
				}
			}

			// prepare for next renderer
			toResolve = &alterInterface{
				scalarData: resolvedValue,
			}
		}

		tmp := toResolve
		// 		tmp := &alterInterface{}
		// 		switch {
		// 		case target.Kind() == reflect.String:
		// 			tmp.scalarData = string(resolvedValue)
		// 		case target.Kind() == reflect.Slice && target.Type().Elem().Kind() == reflect.Uint8:
		// 			tmp.scalarData = resolvedValue
		// 		default:
		// 			// err = yaml.Unmarshal(resolvedValue, tmp)
		// 		}
		//
		// 		if err != nil {
		// 			switch typ := target.Type(); typ {
		// 			case rawInterfaceType, anyObjectMapType:
		// 				// no idea what type is expected, keep it raw
		// 				tmp.mapData = nil
		// 				tmp.sliceData = nil
		// 				tmp.scalarData = string(resolvedValue)
		// 			default:
		// 				// rare case
		// 				return fmt.Errorf(
		// 					"unexpected value type %q, not string, bytes or valid yaml %q: %w",
		// 					typ.String(), resolvedValue, err,
		// 				)
		// 			}
		// 		} else {
		// 			// sometimes go-yaml will parse the input as string when it is not yaml
		// 			// in that case will leave result malformed
		// 			//
		// 			// here we revert that change by checking and resetting scalarData to
		// 			// resolvedValue when it's resolved as string
		// 			switch tmp.scalarData.(type) {
		// 			case string:
		// 				tmp.scalarData = string(resolvedValue)
		// 			case []byte:
		// 				tmp.scalarData = resolvedValue
		// 			case nil, bool, uintptr,
		// 				complex64, complex128,
		// 				float32, float64,
		// 				int, int8, int16, int32, int64,
		// 				uint, uint8, uint16, uint32, uint64:
		// 				// unmarshaled scalar types, do nothing
		// 			case interface{}:
		// 				// TODO: narrow down the interface{} match
		// 				// 		 this case matches all other types
		//
		// 				// map/struct and array/slice types are handled in arrayData and mapData
		// 				// so we don't have to worry about these cases here
		// 				tmp.scalarData = string(resolvedValue)
		// 			}
		// 		}

		if v.isCatchOtherField {
			tmp = &alterInterface{
				mapData: map[string]*alterInterface{
					key.yamlKey: tmp,
				},
			}
		}

		// TODO: currently we always keepOld when the field has tag
		// 		 `rs:"other"`, need to ensure this behavior won't
		// 	     leave inconsistent data

		actualKeepOld := keepOld || v.isCatchOtherField || i != 0
		err = f.unmarshal(key.yamlKey, tmp, target, actualKeepOld)
		if err != nil {
			return fmt.Errorf(
				"failed to unmarshal resolved value to field: %w",
				err,
			)
		}
	}

	return tryResolve(rc, depth-1, target)
}

// resolve user provided data as patch spec
func (f *BaseField) resolvePatchSpec(
	rc RenderingHandler,
	toResolve interface{},
) (
	patchSpec *renderingPatchSpec,
	value interface{},
	err error,
) {
	var patchSpecBytes []byte
	switch t := toResolve.(type) {
	case string:
		patchSpecBytes = []byte(t)
	case []byte:
		patchSpecBytes = t
	default:
		patchSpecBytes, err = yaml.Marshal(toResolve)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"failed to marshal renderer data for patch spec: %w",
				err,
			)
		}
	}

	patchSpec = Init(&renderingPatchSpec{}, f.ifaceTypeHandler).(*renderingPatchSpec)
	err = yaml.Unmarshal(patchSpecBytes, patchSpec)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to unmarshal patch spec %q: %w",
			string(patchSpecBytes), err,
		)
	}

	err = patchSpec.ResolveFields(rc, -1)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to resolve patch spec %q: %w",
			string(patchSpecBytes), err,
		)
	}

	if patchSpec.Value != nil {
		value = patchSpec.Value.NormalizedValue()
	}

	return patchSpec, value, nil
}

func (f *BaseField) isCatchOtherField(yamlKey string) bool {
	if f.catchOtherFields == nil {
		return false
	}

	_, ok := f.catchOtherFields[yamlKey]
	return ok
}

type TypeHint int8

const (
	TypeHintNone TypeHint = iota
	TypeHintStr
	TypeHintBytes
	TypeHintObjects
	TypeHintMap
	TypeHintInt
	TypeHintFloat
)

func (h TypeHint) String() string {
	switch h {
	case TypeHintNone:
		return ""
	case TypeHintStr:
		return "str"
	case TypeHintBytes:
		return "[]byte"
	case TypeHintObjects:
		return "[]obj"
	case TypeHintMap:
		return "map"
	case TypeHintInt:
		return "int"
	case TypeHintFloat:
		return "float"
	default:
		return "<unknown>"
	}
}

func ParseTypeHint(h string) (TypeHint, error) {
	switch h {
	case "":
		return TypeHintNone, nil
	case "str":
		return TypeHintStr, nil
	case "[]byte":
		return TypeHintBytes, nil
	case "[]obj":
		return TypeHintObjects, nil
	case "map":
		return TypeHintMap, nil
	case "int":
		return TypeHintInt, nil
	case "float":
		return TypeHintFloat, nil
	default:
		return -1, fmt.Errorf("unknown type hint %q", h)
	}
}

// nolint:gocyclo
func applyTypeHint(hint TypeHint, v *alterInterface) (interface{}, error) {
	switch hint {
	case TypeHintNone:
		// no hint, return directly
		return v.NormalizedValue(), nil
	case TypeHintStr:
		if v.originalNode != nil && v.originalNode.Kind == yaml.ScalarNode {
			return v.originalNode.Value, nil
		}

		switch vt := v.NormalizedValue().(type) {
		case []byte:
			return string(vt), nil
		case string:
			return vt, nil
		default:
			bytesV, err := yaml.Marshal(vt)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to marshal %T value as str",
					vt,
				)
			}
			return string(bytesV), nil
		}
	case TypeHintBytes:
		if v.originalNode != nil && v.originalNode.Kind == yaml.ScalarNode {
			return []byte(v.originalNode.Value), nil
		}

		switch vt := v.NormalizedValue().(type) {
		case []byte:
			return vt, nil
		case string:
			return []byte(vt), nil
		default:
			bytesV, err := yaml.Marshal(vt)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to marshal %T value as str",
					vt,
				)
			}
			return bytesV, nil
		}
	case TypeHintObjects:
		nv := v.NormalizedValue()
		switch vt := nv.(type) {
		case []byte:
			var actualValue []interface{}
			err := yaml.Unmarshal(vt, &actualValue)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal bytes %q as object array: %w",
					string(vt), err,
				)
			}
			return actualValue, nil
		case string:
			var actualValue []interface{}
			err := yaml.Unmarshal([]byte(vt), &actualValue)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal string %q as object array: %w",
					vt, err,
				)
			}
			return actualValue, nil
		default:
			switch vk := reflect.ValueOf(nv).Kind(); vk {
			case reflect.Array, reflect.Slice, reflect.Interface, reflect.Ptr:
				return nv, nil
			default:
				return nil, fmt.Errorf(
					"incompatible type %T as object array", vt,
				)
			}
		}
	case TypeHintMap:
		nv := v.NormalizedValue()
		switch vt := nv.(type) {
		case []byte:
			var actualValue map[string]interface{}
			err := yaml.Unmarshal(vt, &actualValue)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal bytes %q as map: %w",
					string(vt), err,
				)
			}
			return actualValue, nil
		case string:
			var actualValue map[string]interface{}
			err := yaml.Unmarshal([]byte(vt), &actualValue)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal string %q as map: %w",
					vt, err,
				)
			}
			return actualValue, nil
		default:
			switch vk := reflect.ValueOf(nv).Kind(); vk {
			case reflect.Map, reflect.Struct, reflect.Interface, reflect.Ptr:
				return nv, nil
			default:
				return nil, fmt.Errorf(
					"incompatible type %T as map", vt,
				)
			}
		}
	case TypeHintInt:
		nv := v.NormalizedValue()
		switch vt := nv.(type) {
		case []byte:
			strV, err := strconv.Unquote(string(vt))
			if err != nil {
				strV = string(vt)
			}

			intV, err := strconv.ParseInt(strV, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal bytes %q as int: %w",
					string(vt), err,
				)
			}

			return int(intV), nil
		case string:
			strV, err := strconv.Unquote(vt)
			if err != nil {
				strV = vt
			}

			intV, err := strconv.ParseInt(strV, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal string %q as int: %w",
					vt, err,
				)
			}

			return int(intV), nil
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64, uintptr:
			return nv, nil
		default:
			rv := reflect.ValueOf(nv)
			switch vk := rv.Kind(); vk {
			case reflect.Float32, reflect.Float64:
				return int(rv.Float()), nil
			default:
				return nil, fmt.Errorf(
					"incompatible type %T as number", vt,
				)
			}
		}
	case TypeHintFloat:
		nv := v.NormalizedValue()
		switch vt := nv.(type) {
		case []byte:
			strV, err := strconv.Unquote(string(vt))
			if err != nil {
				strV = string(vt)
			}

			f64v, err := strconv.ParseFloat(strV, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal bytes %q as float: %w",
					string(vt), err,
				)
			}

			return f64v, nil
		case string:
			strV, err := strconv.Unquote(vt)
			if err != nil {
				strV = vt
			}

			f64v, err := strconv.ParseFloat(strV, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to unmarshal string %q as float: %w",
					vt, err,
				)
			}

			return f64v, nil
		case float32, float64:
			return nv, nil
		default:
			rv := reflect.ValueOf(nv)
			switch vk := rv.Kind(); vk {
			case reflect.Int, reflect.Int8, reflect.Int16,
				reflect.Int32, reflect.Int64:
				return float64(rv.Int()), nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16,
				reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				return float64(rv.Uint()), nil
			default:
				return nil, fmt.Errorf(
					"incompatible type %T as number", vt,
				)
			}
		}
	default:
		return nil, fmt.Errorf("unknown type hint %d", hint)
	}
}
