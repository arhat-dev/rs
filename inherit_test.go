package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

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

		unresolvedCount int
	}{
		{
			name: "Both Empty",

			a: &Foo{},
			b: &Foo{},

			unresolvedCount: 0,
		},
		{
			name: "Empty Inherit Some",

			a: &Foo{},
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				v.addUnresolvedField("data", "test", nil, "Data",
					v._parentValue.FieldByName("Data"), false,
					fakeScalarNode("value-b"),
				)
				return v
			}(),

			unresolvedCount: 1,
		},
		{
			name: "Some Inherit Some",

			a: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				v.addUnresolvedField("data", "test", nil, "Data",
					v._parentValue.FieldByName("Data"), false,
					fakeScalarNode("value-a"),
				)
				return v
			}(),
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				v.addUnresolvedField("data", "test", nil, "Data",
					v._parentValue.FieldByName("Data"), false,
					fakeScalarNode("value-b"),
				)
				return v
			}(),

			unresolvedCount: 1,
		},
		{
			name: "Catch Other Empty Inherit Some With Cache",

			a: &Foo{},
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				v.addUnresolvedField("b", "test", nil, "Data",
					v._parentValue.FieldByName("Data"), true,
					fakeMap(fakeScalarNode("data"), fakeScalarNode("test-data")),
				)

				return v
			}(),

			unresolvedCount: 1,
		},
		{
			name: "Catch Other Some Inherit Some No Cache",

			a: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				v.addUnresolvedField("a", "test", nil, "Data",
					v._parentValue.FieldByName("Data"), true,
					fakeMap(fakeScalarNode("a"), fakeScalarNode("test-data")),
				)
				return v
			}(),
			b: func() *Foo {
				v := Init(&Foo{}, nil).(*Foo)
				v.addUnresolvedField("b", "test", nil, "Data",
					v._parentValue.FieldByName("Data"), true,
					fakeMap(fakeScalarNode("b"), fakeScalarNode("test-data")),
				)

				return v
			}(),

			unresolvedCount: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := Init(test.a, nil).(*Foo)
			b := Init(test.b, nil).(*Foo)

			expectedUnresolvedFields := make(map[string]*unresolvedFieldSpec)
			expectedCatchOtherCache := make(map[string]reflect.Value)
			if test.unresolvedCount > 0 {
				for k, v := range a.unresolvedFields {
					expectedUnresolvedFields[k] = &unresolvedFieldSpec{
						fieldName:      v.fieldName,
						fieldValue:     reflect.Value{},
						rawDataList:    append([]*yaml.Node{}, v.rawDataList...),
						renderers:      append([]*rendererSpec{}, v.renderers...),
						isInlineMapKey: v.isInlineMapKey,
					}
				}
			} else {
				expectedUnresolvedFields = nil
			}

			for k, v := range a.inlineMapCache {
				expectedCatchOtherCache[k] = v
			}

			assert.NoError(t, a.Inherit(&b.BaseField))

			assert.NotEqualValues(t, a._parentValue, b._parentValue)
			assert.Len(t, a.unresolvedFields, test.unresolvedCount)

			for k, v := range a.unresolvedFields {
				// value destionation should be redirected to a
				assert.Equal(t, a._parentValue.FieldByName("Data"), v.fieldValue)

				// reset for assertion
				a.unresolvedFields[k].fieldValue = reflect.Value{}
			}

			for k, v := range b.unresolvedFields {
				// reset for assertion
				if _, ok := expectedUnresolvedFields[k]; ok {
					expectedUnresolvedFields[k].rawDataList = append(
						expectedUnresolvedFields[k].rawDataList,
						v.rawDataList...,
					)
				} else {
					expectedUnresolvedFields[k] = &unresolvedFieldSpec{
						fieldName:      v.fieldName,
						fieldValue:     reflect.Value{},
						rawDataList:    append([]*yaml.Node{}, v.rawDataList...),
						renderers:      append([]*rendererSpec{}, v.renderers...),
						isInlineMapKey: v.isInlineMapKey,
					}
				}
			}

			for k, v := range b.inlineMapCache {
				expectedCatchOtherCache[k] = v
			}

			// if len(expectedCatchOtherCache) == 0 {
			// 	expectedCatchOtherCache = nil
			// }

			assert.EqualValues(t, expectedUnresolvedFields, a.unresolvedFields)
			assert.EqualValues(t, expectedCatchOtherCache, a.inlineMapCache)
		})
	}
}
