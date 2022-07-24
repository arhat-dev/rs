package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (f fieldRef) stripBase() *fieldRef {
	f.base = nil
	return &f
}

func TestBaseField_Inherit(t *testing.T) {
	type Foo struct {
		BaseField

		Data  string            `yaml:"data"`
		Other map[string]string `rs:"other"`
	}

	tests := []struct {
		name string

		a *Foo
		b *Foo

		unresolvedNormalFieldsCount  int
		unresolvedInlineMapItemCount int
	}{
		{
			name: "Both Empty",

			a: &Foo{},
			b: &Foo{},

			unresolvedNormalFieldsCount: 0,
		},
		{
			name: "Empty Inherit Some",

			a: &Foo{},
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				_ = v.addUnresolvedField(
					&fieldRef{
						tagName:    "data",
						fieldName:  "Data",
						fieldValue: v._parentValue.FieldByName("Data"),
					},
					fakeScalarNode("value-b"),
					"data", "test", nil,
				)
				return v
			}(),

			unresolvedNormalFieldsCount: 1,
		},
		{
			name: "Some Inherit Some",

			a: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				_ = v.addUnresolvedField(
					&fieldRef{
						tagName:    "data",
						fieldName:  "Data",
						fieldValue: v._parentValue.FieldByName("Data"),
					},
					fakeScalarNode("value-a"),
					"data", "test", nil,
				)
				return v
			}(),
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				_ = v.addUnresolvedField(
					&fieldRef{
						tagName:    "data",
						fieldName:  "Data",
						fieldValue: v._parentValue.FieldByName("Data"),
					},
					fakeScalarNode("value-b"),
					"data", "test", nil,
				)
				return v
			}(),

			unresolvedNormalFieldsCount: 1,
		},
		{
			name: "Catch Other Empty Inherit Some With Cache",

			a: &Foo{},
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				_ = v.addUnresolvedField(
					&fieldRef{
						tagName:     "data",
						fieldName:   "Data",
						fieldValue:  v._parentValue.FieldByName("Data"),
						isInlineMap: true,
					},
					fakeMapPtr(fakeScalarNode("data"), fakeScalarNode("test-data")),
					"b", "test", nil,
				)

				return v
			}(),

			unresolvedInlineMapItemCount: 1,
		},
		{
			name: "Catch Other Some Inherit Some No Cache",

			a: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				_ = v.addUnresolvedField(
					&fieldRef{
						tagName:     "data",
						fieldName:   "Data",
						fieldValue:  v._parentValue.FieldByName("Data"),
						isInlineMap: true,
					},
					fakeMapPtr(fakeScalarNode("a"), fakeScalarNode("test-data")),
					"a", "test", nil,
				)
				return v
			}(),
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				_ = v.addUnresolvedField(
					&fieldRef{
						tagName:     "data",
						fieldName:   "Data",
						fieldValue:  v._parentValue.FieldByName("Data"),
						isInlineMap: true,
					},
					fakeMapPtr(fakeScalarNode("b"), fakeScalarNode("test-data")),
					"b", "test", nil,
				)

				return v
			}(),

			unresolvedInlineMapItemCount: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := Init(test.a, nil).(*Foo)
			b := Init(test.b, nil).(*Foo)

			expectedUnresolvedFields := make(map[string]unresolvedFieldSpec)
			if test.unresolvedNormalFieldsCount > 0 {
				for k, v := range a.unresolvedNormalFields {
					expectedUnresolvedFields[k] = unresolvedFieldSpec{
						ref:       v.ref.clone(reflect.Value{}).stripBase(),
						rawData:   v.rawData,
						renderers: append([]rendererSpec{}, v.renderers...),
					}
				}
			} else {
				expectedUnresolvedFields = nil
			}

			assert.NoError(t, a.Inherit(&b.BaseField))

			assert.NotEqualValues(t, a._parentValue, b._parentValue)
			assert.Len(t, a.unresolvedNormalFields, test.unresolvedNormalFieldsCount)
			assert.Len(t, a.unresolvedInlineMapItems, test.unresolvedInlineMapItemCount)

			for k, v := range a.unresolvedNormalFields {
				// value destionation should be redirected to a
				assert.Equal(t, a._parentValue.FieldByName("Data"), v.ref.fieldValue)

				// reset for assertion
				s := a.unresolvedNormalFields[k]
				s.ref = s.ref.stripBase()
				s.ref.fieldValue = reflect.Value{}
				a.unresolvedNormalFields[k] = s
			}

			for k, v := range b.unresolvedNormalFields {
				// reset for assertion
				if x, ok := expectedUnresolvedFields[k]; ok {
					x.rawData = v.rawData
					expectedUnresolvedFields[k] = x
				} else {
					expectedUnresolvedFields[k] = unresolvedFieldSpec{
						ref:       v.ref.clone(reflect.Value{}).stripBase(),
						rawData:   v.rawData,
						renderers: append([]rendererSpec{}, v.renderers...),
					}
				}
			}

			assert.EqualValues(t, expectedUnresolvedFields, a.unresolvedNormalFields)
		})
	}
}

func TestBaseField_Inherit_uninitialized(t *testing.T) {
	bf := &BaseField{}
	assert.NoError(t, bf.Inherit(nil))
	assert.NoError(t, bf.Inherit(bf))

	type Other struct {
		BaseField
	}

	other := Init(&Other{}, nil).(*Other)
	assert.Error(t, bf.Inherit(&other.BaseField))
	assert.Error(t, other.Inherit(bf))
}
