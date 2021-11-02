package rs

import (
	"fmt"
	"reflect"
)

// Inherit unresolved fields from another BaseField
// useful when you are merging two structs and want to resolve only once
// to get all unresolved fields set
//
// after a successful function call, f wiil be able to resolve its struct fields
// with unresolved values from b and its own
func (f *BaseField) Inherit(b *BaseField) error {
	if len(b.unresolvedFields) == 0 {
		return nil
	}

	if f.unresolvedFields == nil {
		f.unresolvedFields = make(map[string]*unresolvedFieldSpec)
	}

	for k, v := range b.unresolvedFields {
		existingV, ok := f.unresolvedFields[k]
		if !ok {
			f.unresolvedFields[k] = &unresolvedFieldSpec{
				fieldName:                v.fieldName,
				fieldValue:               f._parentValue.FieldByName(v.fieldName),
				isUnresolvedInlineMapKey: v.isUnresolvedInlineMapKey,
				rawDataList:              v.rawDataList,
				renderers:                v.renderers,
			}

			continue
		}

		switch {
		case existingV.fieldName != v.fieldName,
			existingV.isUnresolvedInlineMapKey != v.isUnresolvedInlineMapKey:
			return fmt.Errorf(
				"rs: invalid field not match, want %q, got %q",
				existingV.fieldName, v.fieldName,
			)
		}

		existingV.rawDataList = append(existingV.rawDataList, v.rawDataList...)
	}

	// TODO: values may disappear
	if len(b.inlineMapCache) != 0 {
		if f.inlineMapCache == nil {
			f.inlineMapCache = make(map[string]reflect.Value)
		}

		for k, v := range b.inlineMapCache {
			f.inlineMapCache[k] = v
		}
	}

	return nil
}
