package rs

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

var (
	fieldInterfaceType = reflect.TypeOf((*Field)(nil)).Elem()
	rawInterfaceType   = reflect.TypeOf((*interface{})(nil)).Elem()
)

// UnmarshalYAML handles parsing of rendering suffix and normal yaml
// unmarshaling
// nolint:gocyclo
func (f *BaseField) UnmarshalYAML(n *yaml.Node) error {
	if atomic.LoadUint32(&f._initialized) == 0 {
		return fmt.Errorf("rs: struct not intialized before unmarshaling")
	}

	parentStructType := f._parentValue.Type()

	if n.Kind != yaml.MappingNode {
		return fmt.Errorf(
			"rs: unexpected non map data %q for struct %q unmarshaling",
			n.Tag, parentStructType.String(),
		)
	}

	type fieldKey struct {
		yamlKey string
	}

	type fieldSpec struct {
		fieldName  string
		fieldValue reflect.Value
		base       *BaseField

		isCatchOther bool
	}

	fields := make(map[fieldKey]*fieldSpec)

	addField := func(
		yamlKey, fieldName string,
		fieldValue reflect.Value,
		base *BaseField,
	) bool {
		key := fieldKey{yamlKey: yamlKey}
		if _, exists := fields[key]; exists {
			return false
		}

		fields[key] = &fieldSpec{
			fieldName: fieldName,

			fieldValue: fieldValue,
			base:       base,
		}
		return true
	}

	getField := func(yamlKey string) *fieldSpec {
		return fields[fieldKey{
			yamlKey: yamlKey,
		}]
	}

	var catchOtherField *fieldSpec

	// get known fields first, skip the first field (the BaseField itself)
	for i := 1; i < parentStructType.NumField(); i++ {
		sf := parentStructType.Field(i)
		if len(sf.PkgPath) != 0 {
			// unexported
			continue
		}

		yTags := strings.Split(sf.Tag.Get("yaml"), ",")
		yamlKey := yTags[0]

		if yamlKey == "-" {
			// ignored explicitly
			continue
		}

		var isInlineField bool
		for _, t := range yTags[1:] {
			switch t {
			case "inline":
				isInlineField = true
			default:
				//
			}
		}

		fieldValue := f._parentValue.Field(i)

		// initialize struct fields accepted by Init(), in case being used later
		// DO NOT USE tryInit, that will only init current field, which will cause
		// error when user try to resolve data not unmarshaled from yaml
		InitRecursively(fieldValue, f.ifaceTypeHandler)

		if !isInlineField {
			if len(yamlKey) == 0 {
				yamlKey = strings.ToLower(sf.Name)
			}

			if !addField(yamlKey, sf.Name, fieldValue, f) {
				return fmt.Errorf(
					"rs: duplicate yaml key %q for %s.%s",
					yamlKey, parentStructType.String(), sf.Name,
				)
			}
		} else {
			// handle inline fields

			// find BaseField in inline field, if exists, let it manage dynamnic fields
			base := f

			var innerBaseF reflect.Value
			switch kind := sf.Type.Kind(); {
			case kind == reflect.Struct:
				// embedded struct
				if sf.Type.Implements(fieldInterfaceType) || reflect.PtrTo(sf.Type).Implements(fieldInterfaceType) {
					innerBaseF = fieldValue.Field(0)
				}
			case kind == reflect.Ptr && sf.Type.Elem().Kind() == reflect.Struct:
				// embedded ptr to struct
				if sf.Type.Implements(fieldInterfaceType) || sf.Type.Elem().Implements(fieldInterfaceType) {
					innerBaseF = fieldValue.Elem().Field(0)
				}
			default:
				return fmt.Errorf(
					"rs: invalid inline tag applied to non struct nor struct pointer field %s.%s",
					parentStructType.String(), sf.Name,
				)
			}

			start := 0
			switch innerBaseF.Kind() {
			case reflect.Struct:
				if innerBaseF.Addr().Type() == baseFieldPtrType {
					base = innerBaseF.Addr().Interface().(*BaseField)
					start = 1
				}
			case reflect.Ptr:
				if innerBaseF.Type() == baseFieldPtrType {
					base = innerBaseF.Interface().(*BaseField)
					start = 1
				}
			}

		inlineFieldLoop:
			for j := start; j < fieldValue.NumField(); j++ {
				innerFt := parentStructType.Field(i).Type.Field(j)

				if len(innerFt.PkgPath) != 0 {
					// unexported
					continue
				}

				innerYamlTags := strings.Split(innerFt.Tag.Get("yaml"), ",")
				innerYamlKey := innerYamlTags[0]

				if innerYamlKey == "-" {
					// ignored explicitly
					continue
				}

				for _, tag := range innerYamlTags[1:] {
					switch tag {
					case "inline":
						// already in a inline struct, do not check inline anymore
						continue inlineFieldLoop
					default:
						// TODO: handle other yaml tags
					}
				}

				if len(innerYamlKey) == 0 {
					innerYamlKey = strings.ToLower(innerFt.Name)
				}

				if !addField(innerYamlKey, innerFt.Name, fieldValue.Field(j), base) {
					return fmt.Errorf(
						"rs: duplicate yaml key %q in inline field %s of %s",
						innerYamlKey, innerFt.Name, parentStructType.String(),
					)
				}
			}
		}

		// rs tag is used to extend yaml tag
		dTags := strings.Split(sf.Tag.Get(TagNameRS), ",")
		for _, t := range dTags {
			switch t {
			case "other":
				// other is used to match unhandled values
				// only supports map[string]Any or []Any

				if catchOtherField != nil {
					return fmt.Errorf(
						"rs: bad field tags in %s: only one map in a struct can have `rs:\"other\"` tag",
						parentStructType.String(),
					)
				}

				catchOtherField = &fieldSpec{
					fieldName:  sf.Name,
					fieldValue: fieldValue,
					base:       f,

					isCatchOther: true,
				}
			case "":
			default:
				return fmt.Errorf("rs: unknown rs tag value %q", t)
			}
		}
	}

	m := make(map[string]*alterInterface)
	err := n.Decode(&m)
	if err != nil {
		return fmt.Errorf(
			"rs: data unmarshal failed for %s: %w",
			parentStructType.String(), err,
		)
	}

	handledYamlValues := make(map[string]struct{})
	// handle rendering suffix
	for rawYamlKey, v := range m {
		suffixAt := strings.LastIndexByte(rawYamlKey, '@')
		yamlKey := rawYamlKey
		if suffixAt == -1 {
			// no rendering suffix, fill value

			if _, ok := handledYamlValues[yamlKey]; ok {
				return fmt.Errorf(
					"rs: duplicate yaml field %q",
					yamlKey,
				)
			}

			handledYamlValues[yamlKey] = struct{}{}

			fSpec := getField(yamlKey)
			if fSpec == nil {
				if catchOtherField == nil {
					return fmt.Errorf(
						"rs: unknown yaml field %q in %s",
						yamlKey, parentStructType.String(),
					)
				}

				fSpec = catchOtherField
				v = &alterInterface{
					mapData: map[string]*alterInterface{
						yamlKey: v,
					},
				}
			}

			err = f.unmarshal(
				yamlKey, v, fSpec.fieldValue, fSpec.isCatchOther,
			)
			if err != nil {
				return fmt.Errorf(
					"rs: failed to unmarshal yaml field %q to struct field %s.%s: %w",
					yamlKey, parentStructType.String(), fSpec.fieldName, err,
				)
			}

			continue
		}

		// has rendering suffix

		yamlKey, suffix := rawYamlKey[:suffixAt], rawYamlKey[suffixAt+1:]

		if _, ok := handledYamlValues[yamlKey]; ok {
			return fmt.Errorf(
				"rs: duplicate yaml field name %q for %s, please note"+
					" rendering suffix won't change the field name",
				yamlKey, parentStructType.String(),
			)
		}

		fSpec := getField(yamlKey)
		if fSpec == nil {
			if catchOtherField == nil {
				return fmt.Errorf(
					"rs: unknown yaml field %q for %s",
					yamlKey, parentStructType.String(),
				)
			}

			fSpec = catchOtherField
			v = &alterInterface{
				mapData: map[string]*alterInterface{
					yamlKey: v,
				},
			}
		}

		// do not unmarshal now, we will render the value
		// and unmarshal when calling ResolveFields

		handledYamlValues[yamlKey] = struct{}{}
		// don't forget the raw name with rendering suffix
		handledYamlValues[rawYamlKey] = struct{}{}

		fSpec.base.addUnresolvedField(
			yamlKey, suffix,
			fSpec.fieldName, fSpec.fieldValue, fSpec.isCatchOther,
			v,
		)
	}

	for k := range handledYamlValues {
		delete(m, k)
	}

	if len(m) == 0 {
		// all values consumed
		return nil
	}

	var unknownFields []string
	for k := range m {
		unknownFields = append(unknownFields, k)
	}
	sort.Strings(unknownFields)

	return fmt.Errorf(
		"rs: unknown yaml fields to %s: %s",
		parentStructType.String(), strings.Join(unknownFields, ", "),
	)
}

func (f *BaseField) unmarshal(
	yamlKey string,
	in *alterInterface,
	outVal reflect.Value,
	keepOld bool,
) error {
	outKind := outVal.Kind()

	if !outVal.IsValid() {
		return fmt.Errorf("invalid unmarshal target")
	}

	// reset to zero value if already set
	if in == nil {
		outVal.Set(reflect.Zero(outVal.Type()))
		return nil
	}

	// we are trying to set value of it, so initialize the pointer when not set before
	for outKind == reflect.Ptr {
		if outVal.IsZero() {
			outVal.Set(reflect.New(outVal.Type().Elem()))
		}

		outVal = outVal.Elem()
		outKind = outVal.Kind()
	}

	switch outKind {
	case reflect.Invalid:
		// no way to know what value we can set
		// NOTE: this should not happen when unmarshaling, shall we panic instead?
		return fmt.Errorf("unexpected nil out value for yaml key %q", yamlKey)
	case reflect.Array:
		return f.unmarshalArray(yamlKey, in, outVal)
	case reflect.Slice:
		return f.unmarshalSlice(yamlKey, in, outVal, keepOld)
	case reflect.Map:
		return f.unmarshalMap(yamlKey, in, outVal, keepOld)
	case reflect.Struct:
		return f.unmarshalRaw(in, outVal)
	case reflect.Interface:
		handled, err := f.unmarshalInterface(yamlKey, in, outVal, keepOld)
		if !handled {
			// using default go-yaml behavior
			return f.unmarshalRaw(in, outVal)
		}

		return err
	default:
		// scalar types
		// val := reflect.ValueOf(in.NormalizedValue())
		// inKind := val.Kind()

		inVal := in.NormalizedValue()
		if inVal == nil {
			outVal.Set(reflect.Zero(outVal.Type()))
			return nil
		}

		inRV := reflect.ValueOf(inVal)
		inKind := inRV.Kind()
		if inKind == outKind {
			// same kind means same scalar type
			outVal.Set(inRV)
			return nil
		}

		switch vt := inVal.(type) {
		case string:
			// input is string, while output is not, maybe it's []byte ([]uint8)
			switch outVal.Interface().(type) {
			case []byte:
				outVal.SetBytes([]byte(vt))
				return nil
			default:
				return f.unmarshalRaw(in, outVal)
			}
		case []byte:
			// input is bytes, while output is not, maybe it's string
			switch outVal.Interface().(type) {
			case string:
				outVal.SetString(string(vt))
				return nil
			default:
				return f.unmarshalRaw(in, outVal)
			}
		default:
			// not compatible, check if assignable
			if err := checkAssignable(yamlKey, inRV, outVal); err != nil {
				return err
			}

			outVal.Set(inRV)
			return nil
		}
	}
}

// unmarshalRaw unmarshals interface{} type value to outVal
func (f *BaseField) unmarshalRaw(in *alterInterface, outVal reflect.Value) error {
	tryInit(outVal, f.ifaceTypeHandler)

	dataBytes, err := yaml.Marshal(in)
	if err != nil {
		return fmt.Errorf(
			"rs: failed to marshal back yaml for %q: %w",
			outVal.String(), err,
		)
	}

	var out interface{}
	if outVal.Kind() != reflect.Ptr {
		out = outVal.Addr().Interface()
	} else {
		out = outVal.Interface()
	}

	err = yaml.Unmarshal(dataBytes, out)
	if err != nil {
		return err
	}

	return nil
}

// unmarshalInterface handles interface type creation
func (f *BaseField) unmarshalInterface(
	yamlKey string,
	in *alterInterface,
	outVal reflect.Value,
	keepOld bool,
) (bool, error) {
	if f.ifaceTypeHandler == nil {
		// use default behavior for interface{} types
		return false, nil
	}

	fVal, err := f.ifaceTypeHandler.Create(outVal.Type(), yamlKey)
	if err != nil {
		if errors.Is(err, ErrInterfaceTypeNotHandled) && outVal.Type() == rawInterfaceType {
			// no type information provided, decode using go-yaml directly
			return false, nil
		}

		return true, fmt.Errorf(
			"failed to create interface field: %w",
			err,
		)
	}

	val := reflect.ValueOf(fVal)

	if err := checkAssignable(yamlKey, val, outVal); err != nil {
		return true, err
	}

	if outVal.CanSet() {
		outVal.Set(val)
	} else {
		outVal.Elem().Set(val)
	}

	// DO NOT use outVal directly, which will always match reflect.Interface
	return true, f.unmarshal(yamlKey, in, val, keepOld)
}

func (f *BaseField) unmarshalArray(yamlKey string, in *alterInterface, outVal reflect.Value) error {
	if in.sliceData == nil && in.HasValue() {
		return fmt.Errorf(
			"unexpected non array data of yaml field %q for %s",
			yamlKey, outVal.Type().String(),
		)
	}

	size := len(in.sliceData)
	if size != outVal.Len() {
		return fmt.Errorf(
			"array size not match for %s: want %d got %d",
			outVal.Type().String(), outVal.Len(), size,
		)
	}

	for i := 0; i < size; i++ {
		itemVal := outVal.Index(i)

		err := f.unmarshal(
			yamlKey, in.sliceData[i], itemVal,
			// always drop existing inner data
			// (actually doesn't matter since it's new)
			false,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to unmarshal #%d array item of yaml field %q for %s: %w",
				i, yamlKey, outVal.Type().String(), err,
			)
		}
	}

	return nil
}

func (f *BaseField) unmarshalSlice(yamlKey string, in *alterInterface, outVal reflect.Value, keepOld bool) error {
	if in.sliceData == nil && in.HasValue() {
		return fmt.Errorf(
			"unexpected non slice data of yaml field %q for %s",
			yamlKey, outVal.Type().String(),
		)
	}

	size := len(in.sliceData)
	sliceVal := reflect.MakeSlice(outVal.Type(), size, size)

	for i := 0; i < size; i++ {
		itemVal := sliceVal.Index(i)

		err := f.unmarshal(
			yamlKey, in.sliceData[i], itemVal,
			// always drop existing inner data
			// (actually doesn't matter since it's new)
			false,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to unmarshal #%d slice item of yaml field %q for %s: %w",
				i, yamlKey, outVal.Type().String(), err,
			)
		}
	}

	if err := checkAssignable(yamlKey, sliceVal, outVal); err != nil {
		return err
	}

	if outVal.IsZero() || !keepOld {
		outVal.Set(sliceVal)
	} else {
		outVal.Set(reflect.AppendSlice(outVal, sliceVal))
	}

	return nil
}

func (f *BaseField) unmarshalMap(yamlKey string, in *alterInterface, outVal reflect.Value, keepOld bool) error {
	if in.mapData == nil && in.HasValue() {
		return fmt.Errorf(
			"unexpected non map data of yaml field %q for %s",
			yamlKey, outVal.Type().String(),
		)
	}

	// map key MUST be string
	if outVal.IsNil() || !keepOld {
		outVal.Set(reflect.MakeMap(outVal.Type()))
	}

	valType := outVal.Type().Elem()

	isCatchOtherField := f.isCatchOtherField(yamlKey)
	if isCatchOtherField {
		if f.catchOtherCache == nil {
			f.catchOtherCache = make(map[string]reflect.Value)
		}
	}

	for k, v := range in.mapData {
		// since indexed map value is not addressable
		// we have to keep the original data in BaseField cache

		// TODO: test keepOld behavior with cached data

		var val reflect.Value

		if isCatchOtherField {
			cachedData, ok := f.catchOtherCache[k]
			if ok {
				val = cachedData
			} else {
				val = reflect.New(valType)
				f.catchOtherCache[k] = val
			}
		} else {
			val = reflect.New(valType)
		}

		err := f.unmarshal(
			// use k rather than `yamlKey`
			// because it can be the field catching other
			// (field tag: `rs:"other"`)
			k, v, val, keepOld,
		)
		if err != nil {
			return fmt.Errorf("failed to unmarshal map value %s for key %q: %w",
				valType.String(), k, err,
			)
		}

		outVal.SetMapIndex(reflect.ValueOf(k), val.Elem())
	}

	return nil
}

func checkAssignable(yamlKey string, in, out reflect.Value) error {
	if !in.Type().AssignableTo(out.Type()) {
		return fmt.Errorf(
			"unexpected value of yaml field %q: want %q, got %q",
			yamlKey, out.Type().String(), in.Type().String(),
		)
	}

	return nil
}