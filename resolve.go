package rs

import (
	"fmt"
	"reflect"
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

			// 			// a type is implied when renderer has a `?<type>` in the end
			// 			// and before the patch suffix `!`
			// 			var typeHint string
			// 			if idx := strings.LastIndexByte(renderer, '?'); idx > 0 {
			// 				typeHint = renderer[idx+1:]
			// 				renderer = renderer[:idx]
			// 			}
			//
			// 			switch {
			// 			case typeHint == "str":
			// 			case toResolve != nil:
			// 			}

			// toResolve can only be nil when patch value is not set
			resolvedValue, err = rc.RenderYaml(renderer.name, toResolve.NormalizedValue())
			if err != nil {
				return fmt.Errorf(
					"renderer %q failed to render value: %w",
					renderer.name, err,
				)
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
