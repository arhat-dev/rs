package rs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBaseField_MarshalYAML(t *testing.T) {
	valTrue := true
	valFalse := false
	pTrue := &valTrue
	pFalse := &valFalse

	type (
		Foo              struct{ Foo string }
		FooWithBaseField struct {
			BaseField
			Foo string
		}
	)

	tests := []struct {
		name string

		data Field

		inputNoRS   string
		inputWithRS string

		equivalent interface{}
	}{
		{
			name: "Empty",
			data: &struct{ BaseField }{},

			inputNoRS:   "",
			inputWithRS: "",

			equivalent: &struct{}{},
		},

		{
			name: "String",
			data: &struct {
				BaseField
				Str string
			}{Str: "str"},

			inputNoRS:   "str: str",
			inputWithRS: "str@echo: str",

			equivalent: &struct{ Str string }{Str: "str"},
		},

		{
			name: "Int",
			data: &struct {
				BaseField
				Int int
			}{Int: 100},

			inputNoRS:   "int: 100",
			inputWithRS: "int@echo: 100",

			equivalent: &struct{ Int int }{Int: 100},
		},

		{
			name: "Float",
			data: &struct {
				BaseField
				Float float64
			}{Float: 10.1},

			inputNoRS:   "float: 10.1",
			inputWithRS: "float@echo: 10.1",

			equivalent: &struct{ Float float64 }{Float: 10.1},
		},

		{
			name: "Bool Ptr with Value",
			data: &struct {
				BaseField
				Bool *bool
			}{Bool: pTrue},

			inputNoRS:   "bool: true",
			inputWithRS: "bool@echo: true",

			equivalent: &struct{ Bool *bool }{Bool: pTrue},
		},

		{
			name: "Bool Ptr No Value",
			data: &struct {
				BaseField
				Bool *bool
			}{Bool: nil},

			inputNoRS:   "bool: null",
			inputWithRS: "bool@echo: ",

			equivalent: &struct{ Bool *bool }{Bool: nil},
		},

		{
			name: "Bool Ptr No Value omitempty",
			data: &struct {
				BaseField
				Bool *bool `yaml:",omitempty"`
			}{Bool: nil},

			inputNoRS:   "bool: ",
			inputWithRS: "bool@echo: null",

			equivalent: &struct {
				Bool *bool `yaml:",omitempty"`
			}{Bool: nil},
		},
		{
			name: "Bool Ptr with Value omitempty",
			data: &struct {
				BaseField
				Bool *bool `yaml:",omitempty"`
			}{Bool: pFalse},

			inputNoRS:   "bool: false",
			inputWithRS: "bool@echo: false",

			equivalent: &struct {
				Bool *bool `yaml:",omitempty"`
			}{Bool: pFalse},
		},

		{
			name: "Slice Nil",
			data: &struct {
				BaseField
				Slice []bool
			}{Slice: nil},

			inputNoRS:   "",
			inputWithRS: "",

			equivalent: &struct {
				Slice []bool
			}{Slice: nil},
		},
		{
			name: "Slice Empty",
			data: &struct {
				BaseField
				Slice []bool
			}{Slice: make([]bool, 0)},

			inputNoRS:   "slice: []",
			inputWithRS: "slice@echo: []",

			equivalent: &struct {
				Slice []bool
			}{Slice: make([]bool, 0)},
		},
		{
			name: "Slice omitempty",
			data: &struct {
				BaseField
				Slice []bool `yaml:",omitempty"`
			}{Slice: make([]bool, 0)},

			inputNoRS:   "slice: []",
			inputWithRS: "slice@echo: []",

			equivalent: &struct {
				Slice []bool `yaml:",omitempty"`
			}{Slice: make([]bool, 0)},
		},
		{
			// to address https://github.com/go-yaml/yaml/issues/362
			name: "Inline Struct",
			data: &struct {
				BaseField
				Struct struct {
					BaseField
					Foo string
				} `yaml:",inline"`
			}{Struct: struct {
				BaseField
				Foo string
			}{Foo: "foo"}},

			inputNoRS:   "foo: foo",
			inputWithRS: "foo@echo: foo",

			equivalent: &struct {
				Struct struct {
					Foo string
				} `yaml:",inline"`
			}{Struct: struct{ Foo string }{Foo: "foo"}},
		},
		{
			// to address https://github.com/go-yaml/yaml/issues/362
			name: "Embedded Inline Struct",
			data: &struct {
				BaseField
				FooWithBaseField `yaml:",inline"`
			}{FooWithBaseField: FooWithBaseField{Foo: "foo"}},

			inputNoRS:   "foo: foo",
			inputWithRS: "foo@echo: foo",

			equivalent: &struct {
				Foo `yaml:",inline"`
			}{Foo: Foo{Foo: "foo"}},
		},

		// 		{
		// 			// to address https://github.com/go-yaml/yaml/issues/362
		// 			name: "Multi level Embedded Inline Struct",
		// 			data: &struct {
		// 				BaseField
		//
		// 				FooWithBaseField `yaml:",inline"`
		// 			}{FooWithBaseField: FooWithBaseField{Foo: "foo"}},
		//
		// 			inputNoRS:   "foo: foo",
		// 			inputWithRS: "foo@echo: foo",
		//
		// 			equivalent: &struct {
		// 				Foo `yaml:",inline"`
		// 			}{Foo: Foo{Foo: "foo"}},
		// 		},

		{
			name: "Interface Nil",
			data: &struct {
				BaseField
				IFace yaml.Marshaler
			}{IFace: nil},

			inputNoRS:   "iface: null",
			inputWithRS: "iface@echo: ",

			equivalent: &struct {
				IFace yaml.Marshaler
			}{IFace: nil},
		},

		{
			name: "Struct Ptr Nil Panic if not check value kind",
			data: &struct {
				BaseField
				Ptr *FooWithBaseField
			}{Ptr: nil},

			inputNoRS:   "ptr: ",
			inputWithRS: "ptr@echo: null",

			equivalent: &struct {
				Ptr *Foo
			}{Ptr: nil},
		},
		{
			name: "Struct Ptr Nil Not Panic When omitempty",
			data: &struct {
				BaseField
				Ptr *FooWithBaseField `yaml:",omitempty"`
			}{Ptr: nil},

			inputNoRS:   "ptr: null",
			inputWithRS: "ptr@echo: ",

			equivalent: &struct {
				Ptr *Foo `yaml:",omitempty"`
			}{Ptr: nil},
		},
		{
			name: "Catch Other Only",
			data: &struct {
				BaseField
				Data map[string]string `rs:"other"`
			}{Data: map[string]string{"a": "b", "c": "d"}},

			inputNoRS:   "{a: b, c: d}",
			inputWithRS: "{a@echo: b, c@echo: d}",

			equivalent: map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "Catch Other",
			data: &struct {
				BaseField
				Bar string `yaml:"bar"`

				Data map[string]string `rs:"other"`
			}{Bar: "foo", Data: map[string]string{"a": "b", "c": "d"}},

			inputNoRS:   "{ bar: foo, a: b, c: d}",
			inputWithRS: "{ bar@echo: foo, a@echo: b, c@echo: d}",

			equivalent: map[string]string{
				"bar": "foo",
				"a":   "b",
				"c":   "d",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedBytes, err := yaml.Marshal(test.equivalent)
			assert.NoError(t, err)

			expected := string(expectedBytes)

			t.Run("Direct Set", func(t *testing.T) {
				ret, err := yaml.Marshal(Init(test.data, nil))
				assert.NoError(t, err)

				t.Log(string(ret))
				assert.EqualValues(t, expected, string(ret))
			})

			newEmptyValue := func() Field {
				return reflect.New(reflect.TypeOf(test.data).Elem()).Interface().(Field)
			}

			t.Run("After Unmarshal", func(t *testing.T) {
				v := Init(newEmptyValue(), nil)
				assert.NoError(t, yaml.Unmarshal([]byte(test.inputNoRS), v))

				ret, err := yaml.Marshal(v)
				assert.NoError(t, err)

				assert.EqualValues(t, expected, string(ret))
			})

			t.Run("After Unmarshal And Resolve", func(t *testing.T) {
				v := Init(newEmptyValue(), nil)
				assert.NoError(t, yaml.Unmarshal([]byte(test.inputWithRS), v))

				assert.NoError(t, v.ResolveFields(&testRenderingHandler{}, -1))

				ret, err := yaml.Marshal(v)
				assert.NoError(t, err)

				assert.EqualValues(t, expected, string(ret))
			})
		})
	}
}
