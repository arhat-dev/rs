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
	if other == nil || other == f {
		return nil
	}

	if !f.initialized() {
		return fmt.Errorf("rs.BaseField.Inherit: self not initialized")
	}

	if !other.initialized() {
		return fmt.Errorf("rs.BaseField.Inherit: incoming target not initialized")
	}

	if len(other.unresolvedNormalFields) != 0 {
		if f.unresolvedNormalFields == nil {
			f.unresolvedNormalFields = make(map[string]*unresolvedFieldSpec)
		}

		for k, v := range other.unresolvedNormalFields {
			existingV, ok := f.unresolvedNormalFields[k]
			if !ok {
				f.addUnresolvedField(
					k,
					"", v.renderers,
					v.fieldName,
					f.normalFields[k].fieldValue,
					v.isInlineMapItem,
					v.rawData,
				)

				continue
			}

			switch {
			case existingV.fieldName != v.fieldName,
				existingV.isInlineMapItem != v.isInlineMapItem:
				return fmt.Errorf(
					"rs: invalid field not match, want %q, got %q",
					existingV.fieldName, v.fieldName,
				)
			}

			existingV.rawData = v.rawData
		}
	}

	// here we do not merge rawData for items sharing the same key since their rendering suffix
	// may differ from each other

	if len(other.unresolvedInlineMapItems) != 0 {
		if f.unresolvedInlineMapItems == nil {
			f.unresolvedInlineMapItems = make(map[string][]*unresolvedFieldSpec)
		}

		for k, list := range other.unresolvedInlineMapItems {
			for _, v := range list {
				f.unresolvedInlineMapItems[k] = append(f.unresolvedInlineMapItems[k], &unresolvedFieldSpec{
					fieldName:  v.fieldName,
					fieldValue: f.inlineMap.fieldValue,
					rawData:    v.rawData,
					renderers:  v.renderers,

					isInlineMapItem: true,
				})
			}
		}
	}

	return nil
}
