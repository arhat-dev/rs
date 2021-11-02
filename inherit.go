package rs

import (
	"fmt"
)

// Inherit unresolved fields from another BaseField
// useful when you are merging two structs and want to resolve only once
// to get all unresolved fields set
//
// after a successful function call, f wiil be able to resolve its struct fields
// with unresolved values from b and its own
func (f *BaseField) Inherit(other *BaseField) error {
	if other == nil {
		return nil
	}

	if !f.initialized() {
		return fmt.Errorf("rs.BaseField.Inherit: self not initialized")
	}

	if !other.initialized() {
		return fmt.Errorf("rs.BaseField.Inherit: incoming target not initialized")
	}

	if len(other.unresolvedFields) != 0 {
		if f.unresolvedFields == nil {
			f.unresolvedFields = make(map[string]*unresolvedFieldSpec)
		}

		for k, v := range other.unresolvedFields {
			existingV, ok := f.unresolvedFields[k]
			if !ok {
				f.addUnresolvedField(
					k,
					"", v.renderers,
					v.fieldName,
					f._parentValue.FieldByName(v.fieldName),
					v.isInlineMapKey,
					v.rawDataList...,
				)

				continue
			}

			switch {
			case existingV.fieldName != v.fieldName,
				existingV.isInlineMapKey != v.isInlineMapKey:
				return fmt.Errorf(
					"rs: invalid field not match, want %q, got %q",
					existingV.fieldName, v.fieldName,
				)
			}

			existingV.rawDataList = append(existingV.rawDataList, v.rawDataList...)
		}
	}

	// TODO: values may disappear
	if len(other.inlineMapCache) != 0 {
		if f.inlineMapCache == nil {
			return fmt.Errorf("incompatible type with")
		}

		for k, v := range other.inlineMapCache {
			f.inlineMapCache[k] = v
		}
	}

	return nil
}
